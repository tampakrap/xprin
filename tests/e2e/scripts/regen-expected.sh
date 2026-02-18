#!/bin/bash
# regen-expected.sh - Regenerate e2e expected output files (run inside container)
#
# Behavior is driven by the following environment variables:
#
#   GENERATE=true
#     Find crossplane in PATH -> get version -> XP_MAJOR (same as run.sh) -> output suffix
#     (.v1.output or .output) -> run one generate pass.
#
#   CLEANUP=true
#     Run cleanup (remove redundant .v1.output identical to .output). Can be standalone or
#     after both passes when CROSSPLANE_V1+V2 are set.
#
#   CROSSPLANE_V1 and CROSSPLANE_V2 (paths)
#     Run generate for V1 (.v1.output), then for V2 (.output). Run cleanup only if CLEANUP=true.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
EXPECTED_DIR="${SCRIPT_DIR}/../expected"
TESTCASES_FILE="${SCRIPT_DIR}/testcases.sh"
NORMALIZE_SCRIPT="${SCRIPT_DIR}/normalize.sh"
GEN_INVALID_TESTS_SCRIPT="${SCRIPT_DIR}/gen-invalid-tests.sh"
E2E_TESTS_DIR="${PROJECT_ROOT}/examples/mytests/0_e2e"
XPRIN_BIN="${XPRIN_BIN:-${PROJECT_ROOT}/xprin}"

cd "${PROJECT_ROOT}"

# Clean up generated_*.yaml on exit (created by gen-invalid-tests.sh)
trap 'rm -f "${E2E_TESTS_DIR}"/generated_*.yaml' EXIT

# Detect Crossplane major version (1 or 2) from binary - same logic as run.sh
xp_major_from_binary() {
    local bin="$1"
    local ver
    ver="$("${bin}" version --client 2>/dev/null | cut -d':' -f2 | xargs || true)"
    if [[ "${ver}" == v1.* ]]; then
        echo 1
    elif [[ "${ver}" == v2.* ]]; then
        echo 2
    else
        echo 2
    fi
}

# Map XP_MAJOR to expected output suffix: .v1.output or .output
expected_suffix_for_xp_major() {
    [ "$1" = "1" ] && echo ".v1.output" || echo ".output"
}

run_pass() {
    local suffix="$1"
    echo "Regenerating expected outputs (suffix=${suffix}) into ${EXPECTED_DIR}"
    echo ""
    for test_var in "${TEST_CASES[@]}"; do
        test_id="${test_var#testcase_}"
        test_args="${!test_var}"

        if [ -z "${test_id}" ] || [ -z "${test_args}" ]; then
            echo "Invalid test case entry: ${test_var}"
            exit 1
        fi

        echo "  testcase_${test_id}..."
        read -ra cmd_args <<< "${test_args}"

        ACTUAL_OUTPUT="$(mktemp)"
        set +e
        "${XPRIN_BIN}" test "${cmd_args[@]}" > "${ACTUAL_OUTPUT}" 2>&1
        set -e

        "${NORMALIZE_SCRIPT}" "${ACTUAL_OUTPUT}" > "${EXPECTED_DIR}/testcase_${test_id}${suffix}"
        rm -f "${ACTUAL_OUTPUT}"
    done
    echo ""
}

run_cleanup() {
    echo "--- Removing redundant .v1.output (identical to .output) ---"
    removed=0
    for v1file in "${EXPECTED_DIR}"/testcase_*.v1.output; do
        [ -f "${v1file}" ] || continue
        base="${v1file%.v1.output}"
        outfile="${base}.output"
        if [ -f "${outfile}" ] && cmp -s "${v1file}" "${outfile}"; then
            rm -f "${v1file}"
            echo "  removed $(basename "${v1file}") (identical to $(basename "${outfile}"))"
            removed=$((removed + 1))
        fi
    done
    [ "${removed}" -eq 0 ] && echo "  none"
    echo ""
}

if [ "${GENERATE:-}" = "true" ] || [ -n "${CROSSPLANE_V1:-}" ] || [ -n "${CROSSPLANE_V2:-}" ]; then
    if [ ! -f "${TESTCASES_FILE}" ]; then
        echo "Test case list not found: ${TESTCASES_FILE}"
        exit 1
    fi
    if [ ! -x "${XPRIN_BIN}" ]; then
        echo "xprin binary not found or not executable: ${XPRIN_BIN}"
        exit 1
    fi
    if [ ! -f "${NORMALIZE_SCRIPT}" ]; then
        echo "Normalize script not found: ${NORMALIZE_SCRIPT}"
        exit 1
    fi
    # shellcheck source=/dev/null
    source "${TESTCASES_FILE}"

    TEST_CASES=($(compgen -v | grep '^testcase_' | grep -v '_exit$' | LC_ALL=C sort))
    if [ "${#TEST_CASES[@]}" -eq 0 ]; then
        echo "No test cases defined in ${TESTCASES_FILE}"
        exit 1
    fi

    export E2E_TESTS_DIR

    # Generate schema-invalid e2e testsuite files so they are not committed.
    "${GEN_INVALID_TESTS_SCRIPT}"

    if [ -n "${GENERATE:-}" ] && ([ -n "${CROSSPLANE_V1:-}" ] || [ -n "${CROSSPLANE_V2:-}" ]); then
        echo "Error: GENERATE and CROSSPLANE_V1/CROSSPLANE_V2 cannot be set at the same time"
        exit 1
    fi

    # --- Both: CROSSPLANE_V1 and CROSSPLANE_V2 (paths) ---
    if [ -n "${CROSSPLANE_V1:-}" ] && [ -n "${CROSSPLANE_V2:-}" ]; then
        if [ ! -x "${CROSSPLANE_V1}" ]; then
            echo "CROSSPLANE_V1 not executable: ${CROSSPLANE_V1}"
            exit 1
        fi
        if [ ! -x "${CROSSPLANE_V2}" ]; then
            echo "CROSSPLANE_V2 not executable: ${CROSSPLANE_V2}"
            exit 1
        fi
        export PATH="$(dirname "${CROSSPLANE_V1}"):${PATH}"
        which "${CROSSPLANE_V1}"
        ${CROSSPLANE_V1} version --client
        run_pass ".v1.output"
        export PATH="$(dirname "${CROSSPLANE_V2}"):${PATH}"
        which "${CROSSPLANE_V2}"
        ${CROSSPLANE_V2} version --client
        run_pass ".output"
        echo "Done. Wrote expected file(s) to ${EXPECTED_DIR}."
    fi

    # --- Single: GENERATE=true, crossplane from PATH ---
    if [ "${GENERATE:-}" = "true" ]; then
        CROSSPLANE_BIN="$(command -v crossplane || true)"
        if [ -z "${CROSSPLANE_BIN}" ] || [ ! -x "${CROSSPLANE_BIN}" ]; then
            echo "crossplane not found or not executable on PATH"
            exit 1
        fi
        SUFFIX="$(expected_suffix_for_xp_major "$(xp_major_from_binary "${CROSSPLANE_BIN}")")"
        export PATH="$(dirname "${CROSSPLANE_BIN}"):${PATH}"
        which crossplane
        crossplane version --client
        run_pass "${SUFFIX}"
        echo "Done. Wrote expected file(s) to ${EXPECTED_DIR}."
    fi
fi

if [ "${CLEANUP:-}" = "true" ]; then
    run_cleanup
fi
