// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package liquid

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestForeachOptionType(t *testing.T) {
	liquidOptionTypes := make(map[reflect.Type]struct{})
	getOptionTypesRecursively(t, reflect.ValueOf(ServiceInfo{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(ServiceUsageRequest{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(ServiceUsageReport{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(ServiceCapacityRequest{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(ServiceCapacityReport{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(ServiceQuotaRequest{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(CommitmentChangeRequest{}).Type(), liquidOptionTypes)
	getOptionTypesRecursively(t, reflect.ValueOf(CommitmentChangeResponse{}).Type(), liquidOptionTypes)

	registeredOptionTypes := ForeachOptionType(getOptionType)
	if slices.Contains(registeredOptionTypes, nil) {
		t.Error("ForeachOptionTypeInLIQUID contains values that are not of type github.com/majewsky/gg/option.Option")
	}
	for optionType := range liquidOptionTypes {
		if !slices.Contains(registeredOptionTypes, optionType) {
			t.Errorf("compare option missing for type Option[%s]", optionType)
		}
	}
}

func getOptionTypesRecursively(t *testing.T, rt reflect.Type, optionTypes map[reflect.Type]struct{}) {
	t.Helper()

	switch rt.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Array, reflect.Map:
		getOptionTypesRecursively(t, rt.Elem(), optionTypes)
	case reflect.Struct:
		if strings.HasPrefix(rt.Name(), "Option") {
			field, ok := rt.FieldByName("value")
			if !ok {
				t.Errorf("expected type majewsky/gg/option.Option with field 'value' but got %q", rt.Name())
				return
			}
			optionTypes[field.Type] = struct{}{}
			getOptionTypesRecursively(t, field.Type, optionTypes)
		} else {
			for idx := range rt.NumField() {
				f := rt.Field(idx)
				getOptionTypesRecursively(t, f.Type, optionTypes)
			}
		}
	}
}

func getOptionType(noneValues ...any) reflect.Type {
	if len(noneValues) != 1 || reflect.TypeOf(noneValues[0]).Kind() != reflect.Struct {
		return nil
	}
	field, ok := reflect.TypeOf(noneValues[0]).FieldByName("value")
	if !ok {
		return nil
	}
	return field.Type
}
