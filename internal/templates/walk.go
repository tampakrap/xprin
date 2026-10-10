/*
Copyright 2026 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

import (
	"fmt"
	"reflect"
	"strings"
)

// WalkStrings calls visit for every string reachable from v: field values, slice elements and map
// keys and values. The path tells where the string is, with the names that the fields have in
// YAML (from their json tags), as in `tests[2].inputs.xr` or `common.hooks.pre-test[0].run`. A map
// key has the path of its value, followed by " (key)".
func WalkStrings(v any, visit func(path, s string)) {
	walkStrings(reflect.ValueOf(v), "", visit)
}

//nolint:exhaustive // only the kinds that can contain strings matter
func walkStrings(v reflect.Value, path string, visit func(path, s string)) {
	switch v.Kind() {
	case reflect.String:
		visit(path, v.String())
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			walkStrings(v.Elem(), path, visit)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			name := yamlName(field)
			if name == "" {
				continue
			}

			if path != "" {
				name = path + "." + name
			}

			walkStrings(v.Field(i), name, visit)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			walkStrings(v.Index(i), fmt.Sprintf("%s[%d]", path, i), visit)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprint(iter.Key().Interface())
			valuePath := fmt.Sprintf("%s[%s]", path, key)

			walkStrings(iter.Key(), valuePath+" (key)", visit)
			walkStrings(iter.Value(), valuePath, visit)
		}
	}
}

// yamlName returns the name of a struct field in YAML: the name of its json tag, or the name of
// the field when it has none. It is empty for a field that is left out (`json:"-"`).
func yamlName(field reflect.StructField) string {
	tag, ok := field.Tag.Lookup("json")
	if !ok {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "-" {
		return ""
	}

	if name == "" {
		return field.Name
	}

	return name
}
