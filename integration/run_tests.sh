#!/bin/bash
# Integration tests for troveler. A passing check needs successful execution
# and the expected stdout; stderr must never be treated as result content.
set -u -o pipefail

RESULTS_DIR="${RESULTS_DIR:-/app/results}"
mkdir -p "$RESULTS_DIR" || exit 1
STDERR_FILE=$(mktemp) || exit 1
trap 'rm -f "$STDERR_FILE"' EXIT
export NO_COLOR=1
PASS_COUNT=0
FAIL_COUNT=0
COMMAND_STATUS=0
OUTPUT=""

run_command() {
    COMMAND_STATUS=0
    OUTPUT=$("$@" 2>"$STDERR_FILE") || COMMAND_STATUS=$?
    return "$COMMAND_STATUS"
}

log_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"

    if [ "$status" = "PASS" ]; then
        echo "[PASS] $test_name"
        echo "PASS: $test_name - $message" >> "$RESULTS_DIR/test_results.txt"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        message="$message (command exit $COMMAND_STATUS)"
        if [ -s "$STDERR_FILE" ]; then
            message="$message: $(cat "$STDERR_FILE")"
        fi
        echo "[FAIL] $test_name: $message"
        echo "FAIL: $test_name - $message" >> "$RESULTS_DIR/test_results.txt"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

echo "=========================================="
echo "Troveler Integration Tests"
echo "=========================================="
echo "Starting tests at $(date)"
echo "Test Results - $(date)" > "$RESULTS_DIR/test_results.txt"
echo "==========================================" >> "$RESULTS_DIR/test_results.txt"

echo "--- Test 1: Basic search ---"
if run_command troveler search bat --format json &&
    jq -se 'length == 1 and (.[0] | type == "array" and any(.[]; .slug == "bat" and .name == "bat"))' \
        <<< "$OUTPUT" > /dev/null; then
    log_result "basic_search" "PASS" "Search returned the bat record"
else
    log_result "basic_search" "FAIL" "Search failed or did not return a JSON bat record"
fi

echo "--- Test 2: APK install command display ---"
if run_command troveler install btop --all &&
    grep -Eq '^│[[:space:]]*linux:alpine[[:space:]]*│[[:space:]]*apk add btop[[:space:]]*│$' <<< "$OUTPUT"; then
    log_result "apk_install_display" "PASS" "btop has the Alpine apk add command"
else
    log_result "apk_install_display" "FAIL" "Install display failed or lacks btop Alpine command row"
fi

echo "--- Test 3: APK install execution (btop) ---"
if run_command sudo apk add --no-cache btop && command -v btop > /dev/null; then
    log_result "apk_install_exec" "PASS" "apk succeeded and btop is available"
else
    log_result "apk_install_exec" "FAIL" "apk failed or btop is not available"
fi

echo "--- Test 4: Go install command display ---"
if run_command troveler install diffnav --all &&
    grep -Eq '^│[[:space:]]*go[[:space:]]*│[[:space:]]*go install github.com/dlvhdr/diffnav@[^[:space:]│]+[[:space:]]*│$' \
        <<< "$OUTPUT"; then
    log_result "go_install_display" "PASS" "diffnav has its Go install command"
else
    log_result "go_install_display" "FAIL" "Install display failed or lacks diffnav Go command row"
fi

echo "--- Test 5: Go toolchain verification ---"
if run_command go version && grep -Eq '^go version go[0-9]+\.[0-9]+[^ ]* [^ ]+$' <<< "$OUTPUT"; then
    log_result "go_toolchain" "PASS" "Go toolchain ran successfully"
else
    log_result "go_toolchain" "FAIL" "Go failed or did not report a version"
fi

echo "--- Test 6: Mise integration ---"
if run_command troveler install bat --mise &&
    grep -Eq '^mise use --global cargo:bat(@[^[:space:]]+)?$' <<< "$OUTPUT"; then
    log_result "mise_integration" "PASS" "bat transforms to the cargo mise command"
else
    log_result "mise_integration" "FAIL" "Mise display failed or lacks the bat cargo command"
fi

echo "--- Test 7: Batch install dry run ---"
# Both tools have an Alpine method. Decline sudo, blind skipping and mise;
# no --run flag is used, so this validates planning without executing installs.
if run_command troveler install bat btop --reuse-config true <<< $'n\nn\nn' &&
    grep -Fxq 'Command: apk add bat' <<< "$OUTPUT" &&
    grep -Fxq 'Command: apk add btop' <<< "$OUTPUT" &&
    grep -Fxq '✓ Completed: bat' <<< "$OUTPUT" &&
    grep -Fxq '✓ Completed: btop' <<< "$OUTPUT" &&
    grep -Fxq '  ✓ Completed: 2' <<< "$OUTPUT" &&
    ! grep -Eq '(Failed|Skipped):' <<< "$OUTPUT"; then
    log_result "batch_install" "PASS" "bat and btop each have a completed APK plan"
else
    log_result "batch_install" "FAIL" "Batch failed, skipped a tool or lacks both completed APK plans"
fi

echo "--- Test 8: Search filters ---"
if run_command troveler search 'language=go' --format json &&
    jq -se 'length == 1 and (.[0] | type == "array" and length > 0 and all(.[]; (.language | ascii_downcase) == "go"))' \
        <<< "$OUTPUT" > /dev/null; then
    log_result "search_filters" "PASS" "All filtered records have language Go"
else
    log_result "search_filters" "FAIL" "Filter search failed, returned no records or included another language"
fi

echo "--- Test 9: Search as testuser ---"
# exec returns Troveler's status through su; there is no truncating pipeline.
# shellcheck disable=SC2016 # Expand these paths in testuser's shell, not root's.
if run_command su - testuser -c \
    'export PATH="${XDG_DATA_HOME:-$HOME/.local/share}/mise/shims:$HOME/.local/bin:/app:$PATH"; exec troveler search curl --format json' &&
    jq -se 'length == 1 and (.[0] | type == "array" and any(.[]; .slug == "curl" and .name == "curl"))' \
        <<< "$OUTPUT" > /dev/null; then
    log_result "sudo_user_test" "PASS" "Non-root search succeeded and returned curl"
else
    log_result "sudo_user_test" "FAIL" "Non-root search failed or did not return a JSON curl record"
fi

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Passed: $PASS_COUNT"
echo "Failed: $FAIL_COUNT"
echo "" >> "$RESULTS_DIR/test_results.txt"
echo "Summary: $PASS_COUNT passed, $FAIL_COUNT failed" >> "$RESULTS_DIR/test_results.txt"

if [ "$FAIL_COUNT" -gt 0 ]; then
    echo "Some tests failed!"
    exit 1
fi

echo "All tests passed!"
exit 0
