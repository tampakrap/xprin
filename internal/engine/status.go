/*
Copyright 2025 The Crossplane Authors.

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

package engine

// Status represents the status of a test case or job step (assertion, hook, etc.), including how to display it.
type Status struct {
	Value  string // Canonical value (PASS, FAIL, SKIP, ERROR) for comparison, serialization, and display.
	Symbol string // Display symbol (aligned with crossplane beta validate semantics).
}

// String implements fmt.Stringer so status prints as its canonical value.
func (s Status) String() string {
	return s.Value
}

// StatusPass returns the status for a passed test case or job step.
func StatusPass() Status { return Status{Value: "PASS", Symbol: "[✓]"} }

// StatusFail returns the status for a failed test case or job step.
func StatusFail() Status { return Status{Value: "FAIL", Symbol: "[x]"} }

// StatusSkip returns the status for a skipped test case or job step.
func StatusSkip() Status { return Status{Value: "SKIP", Symbol: "[s]"} }

// StatusError returns the status when a test case or job step could not run.
func StatusError() Status { return Status{Value: "ERROR", Symbol: "[!]"} }
