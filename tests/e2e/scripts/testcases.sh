#!/bin/bash
# testcases.sh - Test case definitions for acceptance tests
#
# This file defines all acceptance test cases. Each test case is a simple string variable
# containing space-separated arguments (test files and flags).
# The test ID is extracted from the variable name (e.g., testcase_001 -> "001")
# Use testcase_<ID>_exit to set an expected non-zero exit code.

# Multiple Successful Files (Non-Verbose)
testcase_001="examples/mytests/1_simple_tests/example1_using-xr_xprin.yaml examples/mytests/1_simple_tests/example2_using-claim_xprin.yaml examples/mytests/2_multiple_testcases/example2_multiple-reconciliation-loops-using-common_xprin.yaml"

# Multiple Successful Files (Verbose)
testcase_002="examples/mytests/1_simple_tests/example1_using-xr_xprin.yaml examples/mytests/1_simple_tests/example2_using-claim_xprin.yaml examples/mytests/2_multiple_testcases/example2_multiple-reconciliation-loops-using-common_xprin.yaml -v"

# Multiple Failures - Combined File (Non-Verbose)
testcase_003="examples/mytests/0_combined/example1_failures_xprin.yaml"
testcase_003_exit=1

# Multiple Failures - Combined File (Verbose)
testcase_004="examples/mytests/0_combined/example1_failures_xprin.yaml -v"
testcase_004_exit=1

# Multiple Failures - Combined File (Verbose, show flags)
testcase_005="examples/mytests/0_combined/example1_failures_xprin.yaml -v --show-render --show-validate --show-hooks --show-assertions"
testcase_005_exit=1

# Successful with hooks/validate/assertions (Verbose, show flags)
testcase_006="examples/mytests/0_combined/example2_success_xprin.yaml -v --show-render --show-validate --show-hooks --show-assertions"

# Test with Chained Outputs
testcase_007="examples/mytests/5_chained_tests/example1_chained-test-outputs_xprin.yaml -v --show-render --show-validate"

# Cross-Composition Chaining
testcase_008="examples/mytests/5_chained_tests/example2_cross-composition-chaining_xprin.yaml -v --show-render --show-validate"
