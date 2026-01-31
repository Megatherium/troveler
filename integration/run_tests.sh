#!/bin/bash
# Integration tests for troveler

RESULTS_DIR="/app/results"
PASS_COUNT=0
FAIL_COUNT=0

log_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    if [ "$status" = "PASS" ]; then
        echo "[PASS] $test_name"
        echo "PASS: $test_name - $message" >> "$RESULTS_DIR/test_results.txt"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "[FAIL] $test_name: $message"
        echo "FAIL: $test_name - $message" >> "$RESULTS_DIR/test_results.txt"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

echo "=========================================="
echo "Troveler Integration Tests"
echo "=========================================="
echo ""
echo "Starting tests at $(date)"
echo ""

# Initialize results file
echo "Test Results - $(date)" > "$RESULTS_DIR/test_results.txt"
echo "==========================================" >> "$RESULTS_DIR/test_results.txt"

# ==========================================
# Test 1: Basic search functionality
# ==========================================
echo "--- Test 1: Basic search ---"
if troveler search bat 2>&1 | grep -q "bat"; then
    log_result "basic_search" "PASS" "Search for 'bat' returned results"
else
    log_result "basic_search" "FAIL" "Search for 'bat' did not return expected results"
fi

# ==========================================
# Test 2: Install command display (apk)
# ==========================================
echo "--- Test 2: APK install command display ---"
OUTPUT=$(troveler install btop --all 2>&1)
if echo "$OUTPUT" | grep -q "apk"; then
    log_result "apk_install_display" "PASS" "btop shows apk install option"
else
    log_result "apk_install_display" "FAIL" "btop does not show apk install option"
fi

# ==========================================
# Test 3: Actual APK install with sudo
# ==========================================
echo "--- Test 3: APK install execution (btop) ---"
if sudo apk add --no-cache btop 2>&1; then
    if command -v btop &>/dev/null; then
        log_result "apk_install_exec" "PASS" "btop installed successfully via apk"
    else
        log_result "apk_install_exec" "FAIL" "btop not found after apk install"
    fi
else
    log_result "apk_install_exec" "FAIL" "apk install command failed"
fi

# ==========================================
# Test 4: Go install command display
# ==========================================
echo "--- Test 4: Go install command display ---"
OUTPUT=$(troveler install diffnav --all 2>&1)
if echo "$OUTPUT" | grep -q "go install"; then
    log_result "go_install_display" "PASS" "diffnav shows go install option"
else
    log_result "go_install_display" "FAIL" "diffnav does not show go install option"
fi

# ==========================================
# Test 5: Go toolchain via mise
# ==========================================
echo "--- Test 5: Go toolchain verification ---"
export PATH="/root/.local/share/mise/shims:/root/go/bin:$PATH"

if go version 2>&1 | grep -q "go1"; then
    log_result "go_toolchain" "PASS" "Go toolchain available via mise"
else
    log_result "go_toolchain" "FAIL" "Go toolchain not available"
fi

# ==========================================
# Test 6: Mise install via troveler
# ==========================================
echo "--- Test 6: Mise integration ---"
OUTPUT=$(troveler install bat --mise 2>&1)
if echo "$OUTPUT" | grep -q "mise use"; then
    log_result "mise_integration" "PASS" "Mise command transformation works"
else
    log_result "mise_integration" "FAIL" "Mise command transformation not working"
fi

# ==========================================
# Test 7: Batch install (dry run)
# ==========================================
echo "--- Test 7: Batch install dry run ---"
OUTPUT=$(echo "y" | troveler install bat gomi --reuse-config true 2>&1)
if echo "$OUTPUT" | grep -q "Batch Install"; then
    log_result "batch_install" "PASS" "Batch install mode works"
else
    log_result "batch_install" "FAIL" "Batch install mode not working"
fi

# ==========================================
# Test 8: Search filters
# ==========================================
echo "--- Test 8: Search filters ---"
OUTPUT=$(troveler search "language=go" 2>&1)
if echo "$OUTPUT" | grep -q "go"; then
    log_result "search_filters" "PASS" "Search filters work"
else
    log_result "search_filters" "FAIL" "Search filters not working"
fi

# ==========================================
# Test 9: Sudo flow test (as testuser)
# ==========================================
echo "--- Test 9: Sudo flow (as testuser) ---"
su - testuser -c 'export PATH="/home/testuser/.local/bin:/app:$PATH"; troveler search curl | head -5' > /dev/null 2>&1
if [ $? -eq 0 ]; then
    log_result "sudo_user_test" "PASS" "troveler works as non-root user"
else
    log_result "sudo_user_test" "FAIL" "troveler failed as non-root user"
fi

# ==========================================
# Summary
# ==========================================
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Passed: $PASS_COUNT"
echo "Failed: $FAIL_COUNT"
echo ""

echo "" >> "$RESULTS_DIR/test_results.txt"
echo "==========================================" >> "$RESULTS_DIR/test_results.txt"
echo "Summary: $PASS_COUNT passed, $FAIL_COUNT failed" >> "$RESULTS_DIR/test_results.txt"

# Exit with failure if any tests failed
if [ $FAIL_COUNT -gt 0 ]; then
    echo "Some tests failed!"
    exit 1
fi

echo "All tests passed!"
exit 0
