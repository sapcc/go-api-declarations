// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package opts_test

import (
	"net/url"
	"testing"
	"time"

	"go.xyrillian.de/gg/assert"
	. "go.xyrillian.de/gg/option"
	"go.xyrillian.de/gg/testcapture"

	"github.com/sapcc/go-api-declarations/opts"
)

func expectPanic(t *testing.T, panicMsg string, fn func()) {
	t.Helper()
	result := testcapture.Capture(t.Context(), t.Name(), func(t assert.TestingTB) { fn() })
	assert.Equal(t, result, testcapture.Result{
		Outcome: testcapture.OutcomePanicked,
		Panic:   panicMsg,
	})
}

func expectAnalyzePanic[T any](t *testing.T, panicMsg string) {
	// Errors that are raised by buildStructInfo() should be visible both in
	// BuildQueryString() and ParseQueryString(), because both use buildStructInfo().
	t.Helper()
	expectPanic(t, panicMsg, func() {
		t.Helper()
		var zero T
		_, _ = opts.BuildQueryString(zero) //nolint:errcheck // panics before returning
	})
	expectPanic(t, panicMsg, func() {
		t.Helper()
		_, _ = opts.ParseQueryString[T](url.Values{}) //nolint:errcheck // panics before returning
	})
}

func TestAnalyzePanics(t *testing.T) {
	// non-struct type parameter (panics)
	expectAnalyzePanic[int](t, "options type is not a struct")

	// unexported field
	expectAnalyzePanic[struct {
		data string `q:"data"`
	}](t, `field "data" is unexported and therefore cannot be set`)

	// missing q-tag
	expectAnalyzePanic[struct {
		String string `yaml:"string"`
	}](t, `expected "String" to have a "q:"-tag`)

	// embedded struct with q-tag
	expectAnalyzePanic[struct {
		EmbeddedOpts `q:"embedded"`
	}](t, `expected embedded struct "EmbeddedOpts" to have no "q:"-tag`)

	// duplicate field declarations
	expectAnalyzePanic[struct {
		FooBar    string `q:"foo_bar"`
		FooAndBar string `q:"foo_bar"`
	}](t, `key "foo_bar" is declared on multiple fields`)

	// contradictory field declarations
	expectAnalyzePanic[struct {
		Scope   string   `q:"scope"`
		With    []string `q:"with"`
		WithFoo bool     `q:"with,value:foo"`
		WithBar bool     `q:"with,value:bar"`
	}](t, `key "with" cannot be declared as both a regular field and a value-discriminant field`)

	// invalid flagset declarations
	expectAnalyzePanic[struct {
		WithFoo bool `q:"with,value:foo"`
		WithBar int  `q:"with,value:bar"`
	}](t, `field "WithBar" has "value:" option but is not a bool`)
	expectAnalyzePanic[struct {
		WithFoo bool `q:"with,value:foo,required"`
		WithBar bool `q:"with,value:bar"`
	}](t, `field "WithFoo" cannot have both "value:" and "required" options`)
	expectAnalyzePanic[struct {
		WithFoo bool `q:"with,value:foo,format:Unix"`
		WithBar bool `q:"with,value:bar"`
	}](t, `field "WithFoo" cannot have both "value:" and "format:" options`)
	expectAnalyzePanic[struct {
		WithFoo      bool `q:"with,value:foo"`
		WithFooAgain bool `q:"with,value:foo"`
	}](t, `value "foo" for key "with" is declared on multiple fields`)

	// unknown struct parameter
	type testNested struct {
		String string `q:"string"`
	}
	expectAnalyzePanic[struct {
		Nested testNested `q:"nested"`
	}](t, "structs other than time.Time and option.Option[T] are not supported")

	// unknown time format
	expectAnalyzePanic[struct {
		Time time.Time `q:"time,format:foo"`
	}](t, `unsupported time format "foo"; accepted: DateOnly, DateTime, RFC3339, RFC3339Nano, Unix`)
	expectAnalyzePanic[struct {
		Time time.Time `q:"time"`
	}](t, `time format is missing for field "Time"`)

	// invalid field types
	expectAnalyzePanic[struct {
		Query    string          `q:"query"`
		Callback func(any) error `q:"callback"`
	}](t, `fields of kind func are not supported`)

	expectAnalyzePanic[struct {
		Query  string  `q:"query"`
		Matrix [][]int `q:"matrix"`
	}](t, `slices of type []int are not supported`)
	expectAnalyzePanic[struct {
		Query    string           `q:"query"`
		Settings []Option[string] `q:"settings"`
	}](t, `slices of type option.Option[string] are not supported`)
	expectAnalyzePanic[struct {
		Query    string                    `q:"query"`
		Settings Option[map[string]string] `q:"settings"`
	}](t, `option.Option[T] with structured payload T = map[string]string is not supported`)

	type point struct {
		X int
		Y int
	}
	expectAnalyzePanic[struct {
		Query string           `q:"query"`
		POIs  map[point]string `q:"poi"`
	}](t, `map keys of type opts_test.point are not supported`)
	expectAnalyzePanic[struct {
		Query string           `q:"query"`
		POIs  map[string]point `q:"poi"`
	}](t, `map values of type opts_test.point are not supported`)
}
