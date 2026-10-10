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
	"testing"

	"github.com/stretchr/testify/assert" //nolint:depguard // testify is widely used for testing
)

type walkInner struct {
	Run string `json:"run"`
}

type walkOuter struct {
	Name     string `json:"name"`
	Optional string `json:"optional,omitempty"`
	NoTag    string
	Skipped  string `json:"-"`
	hidden   string
	Hooks    []walkInner       `json:"preTest,omitempty"`
	Files    map[string]string `json:"files,omitempty"`
	Ptr      *walkInner        `json:"ptr,omitempty"`
	NilPtr   *walkInner        `json:"nilPtr,omitempty"`
	Any      any               `json:"any,omitempty"`
	Number   int               `json:"number"`
}

func TestWalkStrings(t *testing.T) {
	v := walkOuter{
		Name:     "a",
		Optional: "b",
		NoTag:    "c",
		Skipped:  "skipped",
		hidden:   "hidden",
		Hooks:    []walkInner{{Run: "echo 1"}, {Run: "echo 2"}},
		Files:    map[string]string{"k.yaml": "v.yaml"},
		Ptr:      &walkInner{Run: "ptr"},
		Any:      "any",
		Number:   1,
	}

	got := map[string]string{}

	WalkStrings(v, func(path, s string) { got[path] = s })

	assert.Equal(t, map[string]string{
		"name":                "a",
		"optional":            "b",
		"NoTag":               "c",
		"preTest[0].run":      "echo 1",
		"preTest[1].run":      "echo 2",
		"files[k.yaml] (key)": "k.yaml",
		"files[k.yaml]":       "v.yaml",
		"ptr.run":             "ptr",
		"any":                 "any",
	}, got)
}

func TestWalkStrings_Pointer(t *testing.T) {
	var got []string

	WalkStrings(&walkOuter{Name: "a"}, func(path, _ string) { got = append(got, path) })

	assert.Contains(t, got, "name")
}

func TestWalkStrings_Nil(t *testing.T) {
	WalkStrings(nil, func(path, s string) { t.Errorf("unexpected string %q at %q", s, path) })
}
