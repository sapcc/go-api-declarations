// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	. "go.xyrillian.de/gg/option"
)

// ParseQueryString parses the query parameters of a *http.Request into opts,
// which must be a non-nil pointer to a struct. Fields are mapped by their "q"
// struct tag, mirroring the behavior of gophercloud.BuildQueryString.
//
// On configuration errors (e.g. non-pointer opts, opts with non-q-tagged fields)
// the function panics, so make sure to test your opts type thoroughly. On user
// errors (unknown query parameter, type conversion failure, missing required field)
// an error is returned. On success nil is returned and
// opts is fully populated.
//
// The parser supports all scalars except complex. Structs of scalars are supported
// and special structs like time.Time. Additionally, it allows Slices (for multiple
// values), option.Option (recommended for optional values) and pointers (as alternative
// to option.Option) of these. Only map[string]string is supported as a map type.
// Some inputs might work but are untested, because they don't make sense to parse to:
//   - option of map (map defaults to nil when missing)
//   - option of slice (slice defaults to nil when missing)
//   - option of struct (all values of the struct default)
//
// Slice fields are supported in two formats (or a mixed form). When
// "format:"-tag is set to "comma-separated" only the latter is accepted:
// Foo []string `q:"foo"`                       // ?foo=a&foo=b  (repeated keys)
// Bar []int    `q:"bar" format:"comma-separated"` // ?bar=1,2,3
//
// map[string]string fields accept either the OpenStack brace notation
// ({'k1':'v1', 'k2':'v2'}) or plain repeated key=value pairs (?m=k1:v1&m=k2:v2).
//
// time.Time fields support the formats RFC3339Nano, RFC3339, DateTime, Date, Unix.
// Likewise to slices, when no "format:"-tag is set, all are accepted:
// baz time.Time `q:"baz" format:"RFC3339"`
func ParseQueryString(r *http.Request, opts any) error {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() != reflect.Pointer || optsValue.IsNil() {
		panic("expected opts to be a non-nil pointer")
	}
	optsValue = optsValue.Elem()
	if optsValue.Kind() != reflect.Struct {
		panic("expected opts to point to a struct")
	}

	// build map of q-tags to field values
	optsType := optsValue.Type()
	type optMeta struct {
		index             int
		format            string // "" or "comma-separated" or time-formats
		requiredAndUnseen bool
	}
	knownOpts := make(map[string]optMeta, optsType.NumField())
	for i := range optsType.NumField() {
		field := optsType.Field(i)
		fieldValue := optsValue.Field(i)
		format := field.Tag.Get("format")
		qTag := field.Tag.Get("q")
		if qTag == "" {
			panic(fmt.Sprintf(`expected %q to have a "q:"-tag`, field.Name))
		}
		optKey := strings.SplitN(qTag, ",", 2)[0]
		if !fieldValue.CanSet() {
			panic(fmt.Sprintf(`field %q is unexported and therefore cannot be set`, optKey))
		}
		// When format:"..." is set, it's a time format.
		if _, ok := timeFormats[format]; format != "" && format != commaSeparatedFormat && format != unixFormat && !ok {
			return fmt.Errorf("unsupported time format %q; accepted: %s, %s", format, supportedHumanReadableFormats, unixFormat)
		}
		knownOpts[optKey] = optMeta{
			index:             i,
			format:            format,
			requiredAndUnseen: field.Tag.Get("required") != ""}
	}

	// iterate the query
	query := r.URL.Query()
	for optKey, rawValues := range query {
		meta, ok := knownOpts[optKey]
		if !ok {
			// Check for dot-notation: struct.field
			if prefix, suffix, found := strings.Cut(optKey, "."); found {
				meta, ok = knownOpts[prefix]
				if !ok {
					return fmt.Errorf("unknown query parameter %q", optKey)
				}
				fieldValue := optsValue.Field(meta.index)
				if err := setStructField(fieldValue, suffix, rawValues); err != nil {
					return fmt.Errorf("invalid value for query parameter %q: %w", optKey, err)
				}
				continue
			}
			return fmt.Errorf("unknown query parameter %q", optKey)
		}
		fieldValue := optsValue.Field(meta.index)
		// When format:"comma-separated" is set, reject repeated query keys.
		if meta.format == commaSeparatedFormat && len(rawValues) > 1 {
			return fmt.Errorf("query parameter %q uses %s format, but was provided as repeated keys", optKey, commaSeparatedFormat)
		}
		// rawValues has multiple entries for ?k=a&k=b, but we also want to accept ?k=a,b (and mixed).
		var values []string
		for _, rawValue := range rawValues {
			rawValue = strings.TrimSpace(rawValue)
			// Don't split OpenStack brace notation map values!
			if strings.HasPrefix(rawValue, "{") && strings.HasSuffix(rawValue, "}") {
				values = append(values, rawValue)
				continue
			}
			for p := range strings.SplitSeq(rawValue, ",") {
				values = append(values, strings.TrimSpace(p))
			}
		}
		// now, just set the value of the field accordingly
		err := setField(fieldValue, values, meta.format)
		if err != nil {
			return fmt.Errorf("invalid value for query parameter %q: %w", optKey, err)
		}
		if !isZero(fieldValue) && meta.requiredAndUnseen {
			meta.requiredAndUnseen = false
			knownOpts[optKey] = meta
		}
	}

	// check that no required fields are missing
	for optKey, meta := range knownOpts {
		if meta.requiredAndUnseen {
			return fmt.Errorf("missing value for query parameter %q", optKey)
		}
	}
	return nil
}

// setStructField handles dot-notation query parameters for nested struct fields.
// It finds the inner field by its q-tag suffix and dispatches to setField.
func setStructField(fv reflect.Value, suffix string, rawValues []string) error {
	// Dereference/allocate pointer
	if fv.Kind() == reflect.Ptr { //nolint:govet // won't inline this...
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}

	if fv.Kind() != reflect.Struct {
		return errors.New("field is not a struct")
	}

	return setInnerStructField(fv, suffix, rawValues)
}

// setInnerStructField finds a field within a struct by its q-tag and sets its value.
func setInnerStructField(fv reflect.Value, suffix string, rawValues []string) error {
	fvType := fv.Type()
	for j := range fvType.NumField() {
		innerField := fvType.Field(j)
		innerQTag := strings.SplitN(innerField.Tag.Get("q"), ",", 2)[0]
		if innerQTag == suffix {
			// Build values the same way as the main loop
			var values []string
			for _, rawValue := range rawValues {
				rawValue = strings.TrimSpace(rawValue)
				if strings.HasPrefix(rawValue, "{") && strings.HasSuffix(rawValue, "}") {
					values = append(values, rawValue)
					continue
				}
				for p := range strings.SplitSeq(rawValue, ",") {
					values = append(values, strings.TrimSpace(p))
				}
			}
			return setField(fv.Field(j), values, innerField.Tag.Get("format"))
		}
	}
	return fmt.Errorf("unknown field %q", suffix)
}

// setField writes values (already comma-expanded) into a single struct field.
// The format parameter carries the value of the "format" struct tag (may be empty).
func setField(fv reflect.Value, values []string, format string) error {
	// unwrap pointer: allocate if nil
	if fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}

	// Handle Option[T] fields: detected by the presence of an IsSome() bool method.
	// Option[T] implements json.Unmarshaler, so we parse values into JSON and unmarshal.
	if _, isOption := fv.Type().MethodByName("IsSome"); isOption {
		if len(values) == 0 {
			return nil // leave as None (zero value)
		}
		innerType := fv.Type().Field(0).Type
		innerKind := innerType.Kind()

		var (
			jsonBytes []byte
			err       error
		)
		switch innerKind {
		case reflect.Slice:
			// Option[[]T] — build a JSON array from all values
			elemKind := innerType.Elem().Kind()
			elemBitSize := int(innerType.Elem().Size() * 8)
			elems := make([]json.RawMessage, len(values))
			for i, v := range values {
				elems[i], err = scalarToJSON(v, elemKind, elemBitSize)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
			}
			jsonBytes, err = json.Marshal(elems)
		case reflect.Struct:
			// Option[time.Time] — parse with parseTime, then marshal to JSON for UnmarshalJSON
			if innerType != reflect.TypeFor[time.Time]() {
				return fmt.Errorf("unsupported Option inner type %s", innerType)
			}
			if len(values) > 1 {
				return fmt.Errorf("expected a single value, got %d", len(values))
			}
			var formatOpt Option[string]
			if format != "" {
				formatOpt = Some(format)
			}
			t, parseErr := parseTime(values[0], formatOpt)
			if parseErr != nil {
				return parseErr
			}
			jsonBytes, err = json.Marshal(t)
		default:
			// Option[scalar] — must be exactly one value
			if len(values) > 1 {
				return fmt.Errorf("expected a single value, got %d", len(values))
			}
			jsonBytes, err = scalarToJSON(values[0], innerKind, int(innerType.Size()*8))
		}
		if err != nil {
			return err
		}
		return fv.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(jsonBytes)
	}

	// some common error checks
	if slices.Contains([]reflect.Kind{reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64}, fv.Kind()) || fv.Type() == reflect.TypeFor[time.Time]() {
		if len(values) == 0 {
			return nil
		}
		if len(values) > 1 {
			return fmt.Errorf("expected a single value, got %d", len(values))
		}
	}

	// Handle time.Time fields: if a format is specified, use only that format;
	// otherwise try multiple formats in order of specificity.
	if fv.Type() == reflect.TypeFor[time.Time]() {
		var formatOpt Option[string]
		if format != "" {
			formatOpt = Some(format)
		}
		t, err := parseTime(values[0], formatOpt)
		if err != nil {
			return err
		}
		fv.Set(reflect.ValueOf(t))
		return nil
	}

	// set scalars
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(values[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(values[0], 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(values[0], 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(values[0], fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(f)
	// set slices
	case reflect.Slice:
		elemType := fv.Type().Elem()
		switch elemType.Kind() {
		case reflect.String:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				sl.Index(i).SetString(v)
			}
			fv.Set(sl)
		case reflect.Bool:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				b, err := strconv.ParseBool(v)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetBool(b)
			}
			fv.Set(sl)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				n, err := strconv.ParseInt(v, 10, elemType.Bits())
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetInt(n)
			}
			fv.Set(sl)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				n, err := strconv.ParseUint(v, 10, elemType.Bits())
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetUint(n)
			}
			fv.Set(sl)
		case reflect.Float32, reflect.Float64:
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				f, err := strconv.ParseFloat(v, elemType.Bits())
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).SetFloat(f)
			}
			fv.Set(sl)
		case reflect.Struct:
			if elemType != reflect.TypeFor[time.Time]() {
				return fmt.Errorf("unsupported slice element type %s", elemType)
			}
			sl := reflect.MakeSlice(fv.Type(), len(values), len(values))
			for i, v := range values {
				var formatOpt Option[string]
				if format != "" {
					formatOpt = Some(format)
				}
				t, err := parseTime(v, formatOpt)
				if err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
				sl.Index(i).Set(reflect.ValueOf(t))
			}
			fv.Set(sl)
		default:
			return fmt.Errorf("unsupported slice element type %s", elemType)
		}
	// set maps
	case reflect.Map:
		if fv.Type().Key().Kind() != reflect.String || fv.Type().Elem().Kind() != reflect.String {
			return errors.New("only map[string]string is supported")
		}
		m, err := parseMapValues(values)
		if err != nil {
			return err
		}
		fv.Set(reflect.ValueOf(m))
	default:
		return fmt.Errorf("unsupported field type %s", fv.Type())
	}
	return nil
}

// parseMapValues parses a list of raw string values into a map[string]string.
// It supports the OpenStack brace notation ({'k1':'v1', 'k2':'v2'}) and plain
// repeated key:value notation (?m=k1:v1&m=k2:v2).
func parseMapValues(values []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		// OpenStack brace notation: {'k1':'v1', 'k2':'v2'}
		if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
			inner := raw[1 : len(raw)-1]
			for pair := range strings.SplitSeq(inner, ",") {
				pair = strings.TrimSpace(pair)
				// strip surrounding single-quotes from key and value
				pair = strings.Trim(pair, "'")
				kv := strings.SplitN(pair, "':'", 2)
				if len(kv) != 2 {
					return nil, fmt.Errorf("invalid map entry %q", pair)
				}
				m[kv[0]] = kv[1]
			}
			continue
		}
		// Plain k:v notation (from repeated ?m=k1:v1&m=k2:v2 or comma-split)
		kv := strings.SplitN(raw, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid map entry %q: expected key:value", raw)
		}
		m[kv[0]] = kv[1]
	}
	return m, nil
}

// scalarToJSON converts a raw string value to its JSON byte representation
// based on the target type's reflect.Kind. bitSize controls the range check
// for integer and float parsing (e.g. 8 for int8, 64 for int64/float64).
func scalarToJSON(s string, kind reflect.Kind, bitSize int) ([]byte, error) {
	switch kind {
	case reflect.String:
		return json.Marshal(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, bitSize)
		if err != nil {
			return nil, err
		}
		return json.Marshal(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, bitSize)
		if err != nil {
			return nil, err
		}
		return json.Marshal(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, bitSize)
		if err != nil {
			return nil, err
		}
		return json.Marshal(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		return json.Marshal(b)
	default:
		return nil, fmt.Errorf("unsupported scalar type %s", kind)
	}
}

// parseTime parses a time string. If format is Some, only that specific format
// is accepted (one of "RFC3339Nano", "RFC3339", "DateTime", "DateOnly", "Unix").
// If format is None, all formats are tried in order of specificity.
func parseTime(s string, format Option[string]) (time.Time, error) {
	if f, ok := format.Unpack(); ok {
		if f == "Unix" {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return time.Time{}, fmt.Errorf("cannot parse %q as %s seconds: %w", s, unixFormat, err)
			}
			return time.Unix(n, 0).UTC(), nil
		}
		// we checked this already when building knownOpts
		layout := timeFormats[f]
		t, err := time.Parse(layout, s)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot parse %q as %s: %w", s, f, err)
		}
		return t, nil
	}

	// No specific format — try all in order
	for _, layout := range timeFormats {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	// Try Unix timestamp (non-negative seconds since epoch)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(n, 0).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as time; accepted: %s, %s", s, supportedHumanReadableFormats, unixFormat)
}
