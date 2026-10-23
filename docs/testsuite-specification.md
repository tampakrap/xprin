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

Every property of the file, by its full path:

<!-- DOCS_START_INDEX -->

- [`common`](#common)
- [`common.inputs`](#inputs)
- [`common.inputs.claim`](#inputs.claim)
- [`common.inputs.xr`](#inputs.xr)
- [`common.inputs.composition`](#inputs.composition)
- [`common.inputs.functions`](#inputs.functions)
- [`common.inputs.crds`](#inputs.crds)
- [`common.inputs.context-files`](#inputs.context-files)
- [`common.inputs.context-values`](#inputs.context-values)
- [`common.inputs.observed-resources`](#inputs.observed-resources)
- [`common.inputs.extra-resources`](#inputs.extra-resources)
- [`common.inputs.function-credentials`](#inputs.function-credentials)
- [`common.patches`](#patches)
- [`common.patches.xrd`](#patches.xrd)
- [`common.patches.connection-secret`](#patches.connection-secret)
- [`common.patches.connection-secret-name`](#patches.connection-secret-name)
- [`common.patches.connection-secret-namespace`](#patches.connection-secret-namespace)
- [`common.hooks`](#hooks)
- [`common.hooks.pre-test`](#hooks.pre-test)
- [`common.hooks.pre-test[*].name`](#hooks.pre-test[*].name)
- [`common.hooks.pre-test[*].run`](#hooks.pre-test[*].run)
- [`common.hooks.post-test`](#hooks.post-test)
- [`common.hooks.post-test[*].name`](#hooks.post-test[*].name)
- [`common.hooks.post-test[*].run`](#hooks.post-test[*].run)
- [`common.assertions`](#assertions)
- [`common.assertions.xprin`](#assertionsxprin)
- [`common.assertions.xprin[*].name`](#assertionsxprin.name)
- [`common.assertions.xprin[*].type`](#assertionsxprin.type)
- [`common.assertions.xprin[*].resource`](#assertionsxprin.resource)
- [`common.assertions.xprin[*].field`](#assertionsxprin.field)
- [`common.assertions.xprin[*].operator`](#assertionsxprin.operator)
- [`common.assertions.xprin[*].value`](#assertionsxprin.value)
- [`common.assertions.diff`](#assertionsdiff-and-assertionsdyff)
- [`common.assertions.diff[*].name`](#assertionsdiff-and-assertionsdyff.name)
- [`common.assertions.diff[*].expected`](#assertionsdiff-and-assertionsdyff.expected)
- [`common.assertions.diff[*].resource`](#assertionsdiff-and-assertionsdyff.resource)
- [`common.assertions.dyff`](#assertionsdiff-and-assertionsdyff)
- [`common.assertions.dyff[*].name`](#assertionsdiff-and-assertionsdyff.name)
- [`common.assertions.dyff[*].expected`](#assertionsdiff-and-assertionsdyff.expected)
- [`common.assertions.dyff[*].resource`](#assertionsdiff-and-assertionsdyff.resource)
- [`tests`](#tests)
- [`tests[*].name`](#tests.name)
- [`tests[*].id`](#tests.id)
- [`tests[*].needs`](#tests.needs)
- [`tests[*].inputs`](#inputs)
- [`tests[*].inputs.claim`](#inputs.claim)
- [`tests[*].inputs.xr`](#inputs.xr)
- [`tests[*].inputs.composition`](#inputs.composition)
- [`tests[*].inputs.functions`](#inputs.functions)
- [`tests[*].inputs.crds`](#inputs.crds)
- [`tests[*].inputs.context-files`](#inputs.context-files)
- [`tests[*].inputs.context-values`](#inputs.context-values)
- [`tests[*].inputs.observed-resources`](#inputs.observed-resources)
- [`tests[*].inputs.extra-resources`](#inputs.extra-resources)
- [`tests[*].inputs.function-credentials`](#inputs.function-credentials)
- [`tests[*].patches`](#patches)
- [`tests[*].patches.xrd`](#patches.xrd)
- [`tests[*].patches.connection-secret`](#patches.connection-secret)
- [`tests[*].patches.connection-secret-name`](#patches.connection-secret-name)
- [`tests[*].patches.connection-secret-namespace`](#patches.connection-secret-namespace)
- [`tests[*].hooks`](#hooks)
- [`tests[*].hooks.pre-test`](#hooks.pre-test)
- [`tests[*].hooks.pre-test[*].name`](#hooks.pre-test[*].name)
- [`tests[*].hooks.pre-test[*].run`](#hooks.pre-test[*].run)
- [`tests[*].hooks.post-test`](#hooks.post-test)
- [`tests[*].hooks.post-test[*].name`](#hooks.post-test[*].name)
- [`tests[*].hooks.post-test[*].run`](#hooks.post-test[*].run)
- [`tests[*].assertions`](#assertions)
- [`tests[*].assertions.xprin`](#assertionsxprin)
- [`tests[*].assertions.xprin[*].name`](#assertionsxprin.name)
- [`tests[*].assertions.xprin[*].type`](#assertionsxprin.type)
- [`tests[*].assertions.xprin[*].resource`](#assertionsxprin.resource)
- [`tests[*].assertions.xprin[*].field`](#assertionsxprin.field)
- [`tests[*].assertions.xprin[*].operator`](#assertionsxprin.operator)
- [`tests[*].assertions.xprin[*].value`](#assertionsxprin.value)
- [`tests[*].assertions.diff`](#assertionsdiff-and-assertionsdyff)
- [`tests[*].assertions.diff[*].name`](#assertionsdiff-and-assertionsdyff.name)
- [`tests[*].assertions.diff[*].expected`](#assertionsdiff-and-assertionsdyff.expected)
- [`tests[*].assertions.diff[*].resource`](#assertionsdiff-and-assertionsdyff.resource)
- [`tests[*].assertions.dyff`](#assertionsdiff-and-assertionsdyff)
- [`tests[*].assertions.dyff[*].name`](#assertionsdiff-and-assertionsdyff.name)
- [`tests[*].assertions.dyff[*].expected`](#assertionsdiff-and-assertionsdyff.expected)
- [`tests[*].assertions.dyff[*].resource`](#assertionsdiff-and-assertionsdyff.resource)

<!-- DOCS_END_INDEX -->

### root

The top level of the file.

<!-- DOCS_START_ROOT -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="root.common"></a>`common` | [`common`](#common) | no | Common config for all tests |
| <a id="root.tests"></a>`tests` | list of [`tests[*]`](#tests) | yes | List of test cases |

<!-- DOCS_END_ROOT -->

### common

Shared by every test case. Every property is optional.

<!-- DOCS_START_COMMON -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="common.inputs"></a>`inputs` | [`inputs`](#inputs) | no | Common inputs (composition, Claim/XR, etc.) for all testcases |
| <a id="common.patches"></a>`patches` | [`patches`](#patches) | no | Common XR patching configuration for all testcases |
| <a id="common.hooks"></a>`hooks` | [`hooks`](#hooks) | no | Common hooks for all testcases |
| <a id="common.assertions"></a>`assertions` | [`assertions`](#assertions) | no | Common assertions to validate rendered resources for all testcases |

<!-- DOCS_END_COMMON -->

### `tests[*]`

A test case: an element of `tests`. It has the same `inputs`, `patches`, `hooks` and `assertions` as `common`.

<!-- DOCS_START_TESTS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="tests.name"></a>`name` | `string` | yes | Descriptive name for the testcase |
| <a id="tests.id"></a>`id` | `string` | no | Unique identifier for the testcase |
| <a id="tests.needs"></a>`needs` | list of `string` (at least 1) | no | IDs of the testcases that this one depends on |
| <a id="tests.inputs"></a>`inputs` | [`inputs`](#inputs) | conditional | Inputs of a testcase (Required unless specified in the common inputs) |
| <a id="tests.patches"></a>`patches` | [`patches`](#patches) | no | XR patching configuration |
| <a id="tests.hooks"></a>`hooks` | [`hooks`](#hooks) | no | Execution hooks |
| <a id="tests.assertions"></a>`assertions` | [`assertions`](#assertions) | no | Assertions to validate rendered resources |

<!-- DOCS_END_TESTS -->

### inputs

Used in `common.inputs` and `tests[*].inputs`.

<!-- DOCS_START_INPUTS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="inputs.claim"></a>`claim` | `string` | conditional | Path to Claim file (one of Claim or XR must be set, either in the test case or in the common inputs) |
| <a id="inputs.xr"></a>`xr` | `string` | conditional | Path to XR file (one of Claim or XR must be set, either in the test case or in the common inputs) |
| <a id="inputs.composition"></a>`composition` | `string` | conditional | Path to composition file (Required unless specified in the common inputs) |
| <a id="inputs.functions"></a>`functions` | `string` | conditional | Path to functions file or directory (Required unless specified in the common inputs) |
| <a id="inputs.crds"></a>`crds` | list of `string` | no | Paths to CRD files |
| <a id="inputs.context-files"></a>`context-files` | map of `string` | no | Map of context keys to file paths |
| <a id="inputs.context-values"></a>`context-values` | map of `string` | no | Map of context keys to inline values |
| <a id="inputs.observed-resources"></a>`observed-resources` | `string` | no | Path to observed resources file |
| <a id="inputs.extra-resources"></a>`extra-resources` | `string` | no | Path to extra resources file |
| <a id="inputs.function-credentials"></a>`function-credentials` | `string` | no | Path to function credentials file |

<!-- DOCS_END_INPUTS -->

### patches

Used in `common.patches` and `tests[*].patches`.

<!-- DOCS_START_PATCHES -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="patches.xrd"></a>`xrd` | `string` | no | Path to the XR's or Claim's XRD |
| <a id="patches.connection-secret"></a>`connection-secret` | `boolean` | no | When true, create a connection secret for the XR |
| <a id="patches.connection-secret-name"></a>`connection-secret-name` | `string` | no | Name of the connection secret |
| <a id="patches.connection-secret-namespace"></a>`connection-secret-namespace` | `string` | no | Namespace of the connection secret |

<!-- DOCS_END_PATCHES -->

### hooks

Used in `common.hooks` and `tests[*].hooks`. Each hook is a `name` and a `run`.

<!-- DOCS_START_HOOKS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="hooks.pre-test"></a>`pre-test` | list of `object` | no | Hooks that are executed before the testcase |
| <a id="hooks.pre-test[*].name"></a>`pre-test[*].name` | `string` | no | Descriptive name for the hook |
| <a id="hooks.pre-test[*].run"></a>`pre-test[*].run` | `string` | yes | Command to run |
| <a id="hooks.post-test"></a>`post-test` | list of `object` | no | Hooks that are executed after the testcase |
| <a id="hooks.post-test[*].name"></a>`post-test[*].name` | `string` | no | Descriptive name for the hook |
| <a id="hooks.post-test[*].run"></a>`post-test[*].run` | `string` | yes | Command to run |

<!-- DOCS_END_HOOKS -->

### assertions

Used in `common.assertions` and `tests[*].assertions`.

<!-- DOCS_START_ASSERTIONS -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="assertions.xprin"></a>`xprin` | list of [`assertions.xprin[*]`](#assertionsxprin) | no | xprin assertions (in-process) |
| <a id="assertions.diff"></a>`diff` | list of [`assertions.diff[*]` and `assertions.dyff[*]`](#assertionsdiff-and-assertionsdyff) | no | diff assertions (go-native compare to golden file) |
| <a id="assertions.dyff"></a>`dyff` | list of [`assertions.diff[*]` and `assertions.dyff[*]`](#assertionsdiff-and-assertionsdyff) | no | dyff assertions (dyff between expected and actual) |

<!-- DOCS_END_ASSERTIONS -->

### `assertions.xprin[*]`

<!-- DOCS_START_ASSERTIONS_XPRIN -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="assertionsxprin.name"></a>`name` | `string` | yes | Descriptive name for the assertion |
| <a id="assertionsxprin.type"></a>`type` | `string` | yes | Type of assertion. Allowed values: `Count`, `Exists`, `NotExists`, `FieldType`, `FieldExists`, `FieldNotExists`, `FieldValue` |
| <a id="assertionsxprin.resource"></a>`resource` | `string` | no | Resource identifier for resource-based assertions (format: Kind/Name e.g. "Cluster/platform-aws-rds") |
| <a id="assertionsxprin.field"></a>`field` | `string` | no | Field path for field-based assertions (e.g., "metadata.name") |
| <a id="assertionsxprin.operator"></a>`operator` | `string` | no | Operator for field value assertions (== or is). Allowed values: `==`, `is` |
| <a id="assertionsxprin.value"></a>`value` | any | no | Expected value for the assertion |

<!-- DOCS_END_ASSERTIONS_XPRIN -->

Which of `resource`, `field`, `operator` and `value` an assertion needs depends on its `type`; see [Assertion types (xprin)](assertions.md#assertion-types-xprin).

### `assertions.diff[*]` and `assertions.dyff[*]`

<!-- DOCS_START_ASSERTIONS_GOLDEN -->

| Property | Type | Required | Description |
| :------- | :--- | :------: | :---------- |
| <a id="assertionsdiff-and-assertionsdyff.name"></a>`name` | `string` | yes | Descriptive name for the assertion |
| <a id="assertionsdiff-and-assertionsdyff.expected"></a>`expected` | `string` | yes | Path to golden (expected) file |
| <a id="assertionsdiff-and-assertionsdyff.resource"></a>`resource` | `string` | no | Resource identifier for resource-based assertions (format: Kind/Name e.g. "Cluster/platform-aws-rds") |

<!-- DOCS_END_ASSERTIONS_GOLDEN -->

See [Golden-file assertions (diff and dyff)](assertions.md#golden-file-assertions-diff-and-dyff).

---

**Next Steps:**
- Learn how to assert rendered resources with [Assertions](assertions.md)
