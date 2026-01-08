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

package runner

import (
	"fmt"
	"strings"
	"text/template"
	"testing"

	"github.com/crossplane-contrib/xprin/internal/api"
	"github.com/crossplane-contrib/xprin/internal/engine"
	testexecutionUtils "github.com/crossplane-contrib/xprin/internal/testexecution/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"  //nolint:depguard // testify is widely used for testing
	"github.com/stretchr/testify/require" //nolint:depguard // testify is widely used for testing
)

const (
	testResource1File = "resource1.yaml"
	testResource2File = "resource2.yaml"
)

func TestNewAssertionExecutor(t *testing.T) {
	outputs := &engine.Outputs{
		Rendered: make(map[string]string),
	}

	// Mock renderTemplate function (no-op for this test)
	renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
		return content, nil
	}

	executor := newAssertionExecutor(
		afero.NewMemMapFs(),
		outputs,
		false,
		map[string]string{},
		api.Inputs{},
		map[string]*engine.TestCaseResult{},
		renderTemplate,
	)

	assert.NotNil(t, executor)
	assert.Equal(t, outputs, executor.outputs)
	assert.False(t, executor.debug)

	executorWithDebug := newAssertionExecutor(
		afero.NewMemMapFs(),
		outputs,
		true,
		map[string]string{},
		api.Inputs{},
		map[string]*engine.TestCaseResult{},
		renderTemplate,
	)
	assert.True(t, executorWithDebug.debug)
}

func TestAssertionExecutor_ExecuteAssertions(t *testing.T) {
	t.Run("executes multiple assertions", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resource1Path := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				testResource1File: resource1Path,
			},
		}

		// Create a test resource file
		err := afero.WriteFile(fs, resource1Path, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertions := []api.Assertion{
			{Name: "count-1", Type: "Count", Value: 1},
			{Name: "exists-1", Type: "Exists", Resource: "Pod/test-pod"},
		}

		allResults, failedResults := executor.executeAssertions(assertions)
		assert.Len(t, allResults, 2)
		assert.Equal(t, "count-1", allResults[0].Name)
		assert.Equal(t, "exists-1", allResults[1].Name)
		assert.Empty(t, failedResults)
	})

	t.Run("continues execution when one assertion fails", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resource1Path := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				testResource1File: resource1Path,
			},
		}

		// Create a test resource file
		err := afero.WriteFile(fs, resource1Path, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertions := []api.Assertion{
			{Name: "count-wrong", Type: "Count", Value: 999},              // Will fail
			{Name: "exists-ok", Type: "Exists", Resource: "Pod/test-pod"}, // Will pass
		}

		allResults, failedResults := executor.executeAssertions(assertions)
		assert.Len(t, allResults, 2)
		assert.Equal(t, engine.StatusFail, allResults[0].Status)
		assert.Equal(t, engine.StatusPass, allResults[1].Status)
		assert.Len(t, failedResults, 1)
		assert.Equal(t, "count-wrong", failedResults[0].Name)
	})

	t.Run("handles execution errors gracefully", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertions := []api.Assertion{
			{Name: "invalid-type", Type: "InvalidType", Value: "test"},
		}

		allResults, failedResults := executor.executeAssertions(assertions)
		assert.Len(t, allResults, 1)
		assert.Equal(t, engine.StatusFail, allResults[0].Status)
		assert.Contains(t, allResults[0].Message, "unsupported assertion type")
		assert.Len(t, failedResults, 1)
		assert.Equal(t, "invalid-type", failedResults[0].Name)
	})

	t.Run("handles empty assertions list", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		allResults, failedResults := executor.executeAssertions([]api.Assertion{})
		assert.Empty(t, allResults)
		assert.Empty(t, failedResults)
	})
}

func TestAssertionExecutor_executeCountAssertion(t *testing.T) {
	t.Run("passes when count matches", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resource1Path := testResource1File
		resource2Path := testResource2File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				testResource1File: resource1Path,
				testResource2File: resource2Path,
			},
		}

		// Create test resource files
		err := afero.WriteFile(fs, resource1Path, []byte("apiVersion: v1\nkind: Pod\n"), 0o644)
		require.NoError(t, err)
		err = afero.WriteFile(fs, resource2Path, []byte("apiVersion: v1\nkind: Pod\n"), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "count-test", Type: "Count", Value: 2}
		result, err := executor.executeCountAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
		assert.Contains(t, result.Message, "found 2 resources")
	})

	t.Run("fails when count does not match", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte("apiVersion: v1\nkind: Pod\n"), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "count-test", Type: "Count", Value: 5}
		result, err := executor.executeCountAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "expected 5 resources, got 1")
	})

	t.Run("handles float64 value from YAML", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte("apiVersion: v1\nkind: Pod\n"), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "count-test", Type: "Count", Value: float64(1)}
		result, err := executor.executeCountAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
	})

	t.Run("fails with invalid value type", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "count-test", Type: "Count", Value: "not-a-number"}
		result, err := executor.executeCountAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "count assertion value must be a number")
	})
}

func TestAssertionExecutor_executeExistsAssertion(t *testing.T) {
	t.Run("passes when resource exists by kind and name", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "exists-test", Type: "Exists", Resource: "Pod/test-pod"}
		result, err := executor.executeExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
		assert.Contains(t, result.Message, "found")
	})

	t.Run("fails when resource does not exist", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "exists-test", Type: "Exists", Resource: "Service/my-service"}
		result, err := executor.executeExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "not found")
	})

	t.Run("fails when resource field is missing", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "exists-test", Type: "Exists", Resource: ""}
		result, err := executor.executeExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires resource field")
	})

	t.Run("fails with invalid resource format", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "exists-test", Type: "Exists", Resource: "Pod/name/extra"}
		result, err := executor.executeExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "must be in format")
	})
}

func TestAssertionExecutor_executeNotExistsAssertion(t *testing.T) {
	t.Run("passes when resource does not exist", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "not-exists-test", Type: "NotExists", Resource: "Service/my-service"}
		result, err := executor.executeNotExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
		assert.Contains(t, result.Message, "not found (as expected)")
	})

	t.Run("fails when resource exists", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "not-exists-test", Type: "NotExists", Resource: "Pod/test-pod"}
		result, err := executor.executeNotExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "found (should not exist)")
	})

	t.Run("fails when resource field is missing", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{Name: "not-exists-test", Type: "NotExists", Resource: ""}
		result, err := executor.executeNotExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires resource field")
	})
}

func TestAssertionExecutor_executeFieldTypeAssertion(t *testing.T) {
	t.Run("passes when field type matches", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// YAML numbers are parsed as float64, which becomes "number" type
		assertion := api.Assertion{
			Name:     "field-type-test",
			Type:     "FieldType",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
			Value:    "number",
		}
		result, err := executor.executeFieldTypeAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
	})

	t.Run("fails when field type does not match", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-type-test",
			Type:     "FieldType",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
			Value:    "string",
		}
		result, err := executor.executeFieldTypeAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "expected string")
	})

	t.Run("fails when required fields are missing", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Missing resource
		assertion := api.Assertion{Name: "field-type-test", Type: "FieldType", Field: "spec.replicas", Value: "int"}
		result, err := executor.executeFieldTypeAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires resource field")

		// Missing field
		assertion = api.Assertion{Name: "field-type-test", Type: "FieldType", Resource: "Pod/test", Value: "int"}
		result, err = executor.executeFieldTypeAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires field")
	})
}

func TestAssertionExecutor_executeFieldExistsAssertion(t *testing.T) {
	t.Run("passes when field exists", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-exists-test",
			Type:     "FieldExists",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
		}
		result, err := executor.executeFieldExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
		assert.Contains(t, result.Message, "exists")
	})

	t.Run("fails when field does not exist", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec: {}
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-exists-test",
			Type:     "FieldExists",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
		}
		result, err := executor.executeFieldExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "does not exist")
	})
}

func TestAssertionExecutor_executeFieldNotExistsAssertion(t *testing.T) {
	t.Run("passes when field does not exist", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec: {}
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-not-exists-test",
			Type:     "FieldNotExists",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
		}
		result, err := executor.executeFieldNotExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
		assert.Contains(t, result.Message, "does not exist (as expected)")
	})

	t.Run("fails when field exists", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-not-exists-test",
			Type:     "FieldNotExists",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
		}
		result, err := executor.executeFieldNotExistsAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "exists (should not exist)")
	})
}

func TestAssertionExecutor_executeFieldValueAssertion(t *testing.T) {
	t.Run("passes when field value matches with ==", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-value-test",
			Type:     "FieldValue",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
			Operator: "==",
			Value:    float64(3), // YAML numbers are parsed as float64
		}
		result, err := executor.executeFieldValueAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)
	})

	t.Run("fails when field value does not match with ==", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-value-test",
			Type:     "FieldValue",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
			Operator: "==",
			Value:    float64(5), // YAML numbers are parsed as float64
		}
		result, err := executor.executeFieldValueAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "expected ==")
	})

	t.Run("fails when required fields are missing", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Missing resource
		assertion := api.Assertion{Name: "field-value-test", Type: "FieldValue", Field: "spec.replicas", Operator: "==", Value: float64(3)}
		result, err := executor.executeFieldValueAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires resource field")

		// Missing field
		assertion = api.Assertion{Name: "field-value-test", Type: "FieldValue", Resource: "Pod/test", Operator: "==", Value: float64(3)}
		result, err = executor.executeFieldValueAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires field")

		// Missing operator
		assertion = api.Assertion{Name: "field-value-test", Type: "FieldValue", Resource: "Pod/test", Field: "spec.replicas", Value: 3}
		result, err = executor.executeFieldValueAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires operator field")

		// Missing value
		assertion = api.Assertion{Name: "field-value-test", Type: "FieldValue", Resource: "Pod/test", Field: "spec.replicas", Operator: "=="}
		result, err = executor.executeFieldValueAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "requires value field")
	})

	t.Run("fails with unsupported operator", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		assertion := api.Assertion{
			Name:     "field-value-test",
			Type:     "FieldValue",
			Resource: "Pod/test-pod",
			Field:    "spec.replicas",
			Operator: "unsupported",
			Value:    float64(3),
		}
		result, err := executor.executeFieldValueAssertion(assertion)

		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "unsupported operator")
	})
}

func TestAssertionExecutor_executeAssertion(t *testing.T) {
	t.Run("routes to correct assertion type", func(t *testing.T) {
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (no-op for this test)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			afero.NewMemMapFs(),
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Test Count
		assertion := api.Assertion{Name: "test", Type: "Count", Value: 0}
		result, err := executor.executeAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusPass, result.Status)

		// Test unsupported type
		assertion = api.Assertion{Name: "test", Type: "Unsupported", Value: "test"}
		result, err = executor.executeAssertion(assertion)
		require.NoError(t, err)
		assert.Equal(t, engine.StatusFail, result.Status)
		assert.Contains(t, result.Message, "unsupported assertion type")
	})
}

func TestAssertionExecutor_processTemplateVariables(t *testing.T) {
	t.Run("processes template variables in assertion fields", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		resourcePath := testResource1File
		outputs := &engine.Outputs{
			Rendered: map[string]string{
				"resource1.yaml": resourcePath,
			},
		}

		err := afero.WriteFile(fs, resourcePath, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  replicas: 3
`), 0o644)
		require.NoError(t, err)

		// Mock renderTemplate function that processes templates
		renderTemplate := func(content string, templateContext *templateContext, templateName string) (string, error) {
			tmpl, err := template.New(templateName).Option("missingkey=error").Parse(content)
			if err != nil {
				return "", err
			}

			var buf strings.Builder
			if err := tmpl.Execute(&buf, templateContext); err != nil {
				return "", err
			}

			return buf.String(), nil
		}

		repositories := map[string]string{
			"myrepo": "/path/to/repo",
		}

		inputs := api.Inputs{
			XR:          "my-xr.yaml",
			Composition: "my-composition.yaml",
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			repositories,
			inputs,
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Create assertion with template variables
		assertion := api.Assertion{
			Name:     testexecutionUtils.CreatePlaceholder(".Inputs.XR"),
			Type:     "Exists",
			Resource: testexecutionUtils.CreatePlaceholder(".Inputs.Composition"),
			Field:    "spec.replicas",
		}

		// Process template variables
		processed, err := executor.processTemplateVariables(assertion)
		require.NoError(t, err)

		// Verify template variables were processed
		assert.Equal(t, "my-xr.yaml", processed.Name)
		assert.Equal(t, "my-composition.yaml", processed.Resource)
		assert.Equal(t, "spec.replicas", processed.Field) // No template var, should remain unchanged
	})

	t.Run("processes template variables in value field", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function that processes templates
		renderTemplate := func(content string, templateContext *templateContext, templateName string) (string, error) {
			tmpl, err := template.New(templateName).Option("missingkey=error").Parse(content)
			if err != nil {
				return "", err
			}

			var buf strings.Builder
			if err := tmpl.Execute(&buf, templateContext); err != nil {
				return "", err
			}

			return buf.String(), nil
		}

		inputs := api.Inputs{
			XR: "my-xr.yaml",
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			inputs,
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Create assertion with template variable in value field
		assertion := api.Assertion{
			Name:  "test",
			Type:  "FieldValue",
			Value: testexecutionUtils.CreatePlaceholder(".Inputs.XR"),
		}

		// Process template variables
		processed, err := executor.processTemplateVariables(assertion)
		require.NoError(t, err)

		// Verify template variable was processed
		assert.Equal(t, "my-xr.yaml", processed.Value)
	})

	t.Run("handles assertions without template variables", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function (should not be called)
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return content, nil
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Create assertion without template variables
		assertion := api.Assertion{
			Name:     "test-assertion",
			Type:     "Count",
			Resource: "Pod/test-pod",
			Value:    1,
		}

		// Process template variables
		processed, err := executor.processTemplateVariables(assertion)
		require.NoError(t, err)

		// Verify assertion remains unchanged
		assert.Equal(t, assertion, processed)
	})

	t.Run("handles template processing errors gracefully", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		outputs := &engine.Outputs{
			Rendered: make(map[string]string),
		}

		// Mock renderTemplate function that returns an error
		renderTemplate := func(content string, _ *templateContext, _ string) (string, error) {
			return "", fmt.Errorf("template error")
		}

		executor := newAssertionExecutor(
			fs,
			outputs,
			false,
			map[string]string{},
			api.Inputs{},
			map[string]*engine.TestCaseResult{},
			renderTemplate,
		)

		// Create assertion with template variable
		assertion := api.Assertion{
			Name: testexecutionUtils.CreatePlaceholder(".Inputs.XR"),
			Type: "Count",
		}

		// Process template variables - should return error
		_, err := executor.processTemplateVariables(assertion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to process template")
	})
}
