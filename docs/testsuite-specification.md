# Test Suite Specification

A test suite file defines test cases for xprin. Its name must be `xprin.yaml` or end with `_xprin.yaml` (see [Discovering Testsuite Files](how-it-works.md#discovering-testsuite-files)). For a complete file, see [Getting Started](getting-started.md) and the [examples](../examples/README.md).

This page describes the format, key by key. The tables are generated from the Go types in `internal/api`; for what xprin does with each key, see [How It Works](how-it-works.md).

## Structure

```yaml
common:               # Optional: shared by every test case
  inputs: {}
  patches: {}
  hooks: {}
  assertions: {}
tests:                # Required
- name: ""            # Required
  id: ""
  needs: []
  inputs: {}          # Same keys as common
  patches: {}
  hooks: {}
  assertions: {}
```

## Rules

- **`common` and test cases are merged.** `inputs`, `patches`, `hooks` and `assertions` can be set in `common`, in a test case, or in both; see [Common vs Test-Level Configuration](how-it-works.md#common-vs-test-level-configuration). A property marked **conditional** in the tables is required in the merged result, so it can come from either.
- **`xr` or `claim`, not both.** One of them is required, in the test case or in `common`.
- **`needs` lists the `id`s of other test cases.** A test case can't list itself, the `needs` can't form a cycle, and each one has to be the `id` of a test case in the same file, listed before or after it. A test case that needs another one runs after it, and is skipped if it failed or was skipped. See [Dependencies Between Test Cases](how-it-works.md#dependencies-between-test-cases-needs), and `xprin graph` to see the resulting graph.
- **Paths** are absolute, relative to the test suite file, or built with template variables (`{{ .Repositories.myrepo }}/...`); see [Path Resolution](how-it-works.md#path-resolution) and [Template Variable Expansion](how-it-works.md#template-variable-expansion).
- **Hooks** are shell commands that run before (`pre-test`) and after (`post-test`) the test case; see [Hooks Execution](how-it-works.md#hooks-execution).
- **Assertions** are grouped by engine (`xprin`, `diff`, `dyff`); see [Assertions](assertions.md).

## Reference

### root

The top level of the file.

<!-- DOCS_START_ROOT -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `common` | [`common`](#common) | no | Common config for all tests |
| `tests` | list of [`tests[*]`](#tests) | yes | List of test cases |

<!-- DOCS_END_ROOT -->

### common

Shared by every test case. Every property is optional.

<!-- DOCS_START_COMMON -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `inputs` | [`inputs`](#inputs) | no | Common inputs (composition, Claim/XR, etc.) for all testcases |
| `patches` | [`patches`](#patches) | no | Common XR patching configuration for all testcases |
| `hooks` | [`hooks`](#hooks) | no | Common hooks for all testcases |
| `assertions` | [`assertions`](#assertions) | no | Common assertions to validate rendered resources for all testcases |

<!-- DOCS_END_COMMON -->

### `tests[*]`

A test case: an element of `tests`. It has the same `inputs`, `patches`, `hooks` and `assertions` as `common`.

<!-- DOCS_START_TESTS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `name` | `string` | yes | Descriptive name for the testcase |
| `id` | `string` | no | Unique identifier for the testcase |
| `needs` | list of `string` (at least 1) | no | IDs of the testcases that this one depends on |
| `inputs` | [`inputs`](#inputs) | conditional | Inputs of a testcase (Required unless specified in the common inputs) |
| `patches` | [`patches`](#patches) | no | XR patching configuration |
| `hooks` | [`hooks`](#hooks) | no | Execution hooks |
| `assertions` | [`assertions`](#assertions) | no | Assertions to validate rendered resources |

<!-- DOCS_END_TESTS -->

### inputs

Used in `common.inputs` and `tests[*].inputs`.

<!-- DOCS_START_INPUTS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `claim` | `string` | conditional | Path to Claim file (one of Claim or XR must be set, either in the test case or in the common inputs) |
| `xr` | `string` | conditional | Path to XR file (one of Claim or XR must be set, either in the test case or in the common inputs) |
| `composition` | `string` | conditional | Path to composition file (Required unless specified in the common inputs) |
| `functions` | `string` | conditional | Path to functions file or directory (Required unless specified in the common inputs) |
| `crds` | list of `string` | no | Paths to CRD files |
| `context-files` | map of `string` | no | Map of context keys to file paths |
| `context-values` | map of `string` | no | Map of context keys to inline values |
| `observed-resources` | `string` | no | Path to observed resources file |
| `extra-resources` | `string` | no | Path to extra resources file |
| `function-credentials` | `string` | no | Path to function credentials file |

<!-- DOCS_END_INPUTS -->

### patches

Used in `common.patches` and `tests[*].patches`.

<!-- DOCS_START_PATCHES -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `xrd` | `string` | no | Path to the XR's or Claim's XRD |
| `connection-secret` | `boolean` | no | When true, create a connection secret for the XR |
| `connection-secret-name` | `string` | no | Name of the connection secret |
| `connection-secret-namespace` | `string` | no | Namespace of the connection secret |

<!-- DOCS_END_PATCHES -->

### hooks

Used in `common.hooks` and `tests[*].hooks`. Each hook is a `name` and a `run`.

<!-- DOCS_START_HOOKS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `pre-test` | list of `object` | no | Hooks that are executed before the testcase |
| `pre-test[*].name` | `string` | no | Descriptive name for the hook |
| `pre-test[*].run` | `string` | yes | Command to run |
| `post-test` | list of `object` | no | Hooks that are executed after the testcase |
| `post-test[*].name` | `string` | no | Descriptive name for the hook |
| `post-test[*].run` | `string` | yes | Command to run |

<!-- DOCS_END_HOOKS -->

### assertions

Used in `common.assertions` and `tests[*].assertions`.

<!-- DOCS_START_ASSERTIONS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `xprin` | list of [`assertions.xprin[*]`](#assertionsxprin) | no | xprin assertions (in-process) |
| `diff` | list of [`assertions.diff[*]` and `assertions.dyff[*]`](#assertionsdiff-and-assertionsdyff) | no | diff assertions (go-native compare to golden file) |
| `dyff` | list of [`assertions.diff[*]` and `assertions.dyff[*]`](#assertionsdiff-and-assertionsdyff) | no | dyff assertions (dyff between expected and actual) |

<!-- DOCS_END_ASSERTIONS -->

### `assertions.xprin[*]`

<!-- DOCS_START_ASSERTIONS_XPRIN -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `name` | `string` | yes | Descriptive name for the assertion |
| `type` | `string` | yes | Type of assertion. Allowed values: `Count`, `Exists`, `NotExists`, `FieldType`, `FieldExists`, `FieldNotExists`, `FieldValue` |
| `resource` | `string` | no | Resource identifier for resource-based assertions (format: Kind/Name e.g. "Cluster/platform-aws-rds") |
| `field` | `string` | no | Field path for field-based assertions (e.g., "metadata.name") |
| `operator` | `string` | no | Operator for field value assertions (== or is). Allowed values: `==`, `is` |
| `value` | any | no | Expected value for the assertion |

<!-- DOCS_END_ASSERTIONS_XPRIN -->

Which of `resource`, `field`, `operator` and `value` an assertion needs depends on its `type`; see [Assertion types (xprin)](assertions.md#assertion-types-xprin).

### `assertions.diff[*]` and `assertions.dyff[*]`

<!-- DOCS_START_ASSERTIONS_GOLDEN -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| `name` | `string` | yes | Descriptive name for the assertion |
| `expected` | `string` | yes | Path to golden (expected) file |
| `resource` | `string` | no | Resource identifier for resource-based assertions (format: Kind/Name e.g. "Cluster/platform-aws-rds") |

<!-- DOCS_END_ASSERTIONS_GOLDEN -->

See [Golden-file assertions (diff and dyff)](assertions.md#golden-file-assertions-diff-and-dyff).

---

**Next Steps:**
- Learn how to assert rendered resources with [Assertions](assertions.md)
