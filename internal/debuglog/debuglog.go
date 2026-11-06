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

// Package debuglog prints the --debug lines of a run. The lines of a test case go to a writer, which is stderr, or the
// buffer of a test case that runs at the same time as others, whose lines are then printed with its result so that they
// are not mixed with theirs. The lines of the run that are not of a test case go to stderr.
package debuglog

import (
	"fmt"
	"io"
	"os"

	"github.com/crossplane-contrib/xprin/internal/api"
	"github.com/crossplane-contrib/xprin/internal/templates/placeholder"
)

// Log is where the --debug lines go: to a writer.
type Log struct {
	w io.Writer
}

// New returns the Log that prints to w.
func New(w io.Writer) *Log {
	return &Log{w: w}
}

// Printf prints a --debug line, with the "DEBUG: " prefix.
func (l *Log) Printf(format string, args ...any) {
	fmt.Fprintf(l.w, "DEBUG: "+format, args...) //nolint:errcheck // output function, error handling not practical
}

// Printf prints a --debug line to stderr: the ones of the run that are not of a test case.
func Printf(format string, args ...any) {
	New(os.Stderr).Printf(format, args...)
}

// printPatches prints patches in a consistent format.
func (l *Log) printPatches(patches api.Patches) {
	if patches.HasPatches() {
		l.Printf("  Patches:\n")

		if patches.XRD != "" {
			l.Printf("  - XRD: %s\n", patches.XRD)
		}

		if patches.ConnectionSecret != nil {
			l.Printf("  - Connection Secret: %t\n", *patches.ConnectionSecret)
		}

		if patches.ConnectionSecretName != "" {
			l.Printf("  - Connection Secret Name: %s\n", patches.ConnectionSecretName)
		}

		if patches.ConnectionSecretNamespace != "" {
			l.Printf("  - Connection Secret Namespace: %s\n", patches.ConnectionSecretNamespace)
		}
	}
}

// printInputs prints the content of inputs with the header.
func (l *Log) printInputs(inputs api.Inputs) {
	l.Printf("  Inputs:\n")

	if inputs.XR != "" {
		l.Printf("  - XR: %s\n", inputs.XR)
	}

	if inputs.Claim != "" {
		l.Printf("  - Claim: %s\n", inputs.Claim)
	}

	if inputs.Composition != "" {
		l.Printf("  - Composition: %s\n", inputs.Composition)
	}

	if inputs.Functions != "" {
		l.Printf("  - Functions: %s\n", inputs.Functions)
	}

	if len(inputs.CRDs) > 0 {
		l.Printf("  - CRDs:\n")

		for _, crd := range inputs.CRDs {
			l.Printf("    - %s\n", crd)
		}
	}

	if len(inputs.ContextFiles) > 0 {
		l.Printf("  - Context Files:\n")

		for contextKey, contextFile := range inputs.ContextFiles {
			l.Printf("      %s: %s\n", contextKey, contextFile)
		}
	}

	if len(inputs.ContextValues) > 0 {
		l.Printf("  - Context Values:\n")

		for contextKey, contextValue := range inputs.ContextValues {
			l.Printf("      %s: %s\n", contextKey, contextValue)
		}
	}

	if inputs.ObservedResources != "" {
		l.Printf("  - Observed Resources: %s\n", inputs.ObservedResources)
	}

	if inputs.ExtraResources != "" {
		l.Printf("  - Extra Resources: %s\n", inputs.ExtraResources)
	}

	if inputs.FunctionCredentials != "" {
		l.Printf("  - Function Credentials: %s\n", inputs.FunctionCredentials)
	}
}

// printHooks prints hooks in a consistent format.
func (l *Log) printHooks(hooks api.Hooks) {
	if hooks.HasHooks() {
		l.Printf("  Hooks:\n")

		if hooks.HasPreTestHooks() {
			l.Printf("    Pre-Test:\n")

			for _, hook := range hooks.PreTest {
				if hook.Name != "" {
					l.Printf("    - name: %s\n", hook.Name)
				}
				// Show processed hook command instead of raw with placeholders
				processedCommand := placeholder.Restore(hook.Run)
				l.Printf("      run: %s\n", processedCommand)
			}
		}

		if hooks.HasPostTestHooks() {
			l.Printf("    Post-Test:\n")

			for _, hook := range hooks.PostTest {
				if hook.Name != "" {
					l.Printf("    - name: %s\n", hook.Name)
				}
				// Show processed hook command instead of raw with placeholders
				processedCommand := placeholder.Restore(hook.Run)
				l.Printf("      run: %s\n", processedCommand)
			}
		}
	}
}

// PrintCommon prints debug information for common configuration.
func (l *Log) PrintCommon(common api.Common, header string) {
	l.Printf("%s\n", header)
	l.printPatches(common.Patches)
	l.printInputs(common.Inputs)
	l.printHooks(common.Hooks)
}

// PrintTestCase prints debug information for a test case.
func (l *Log) PrintTestCase(testCase api.TestCase, header string) {
	l.Printf("%s\n", header)
	l.printPatches(testCase.Patches)
	l.printInputs(testCase.Inputs)
	l.printHooks(testCase.Hooks)
}
