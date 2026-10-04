package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These commands exercise the runner's decisions without installing packages,
// accessing a user database, changing users, or reaching the network. jq remains
// real so search output must actually parse and satisfy the JSON predicates.
const fixtureCommand = `#!/bin/bash
set -u
case "${0##*/}" in
    troveler)
        case "$1:$2" in
            search:bat)
                case "$RUNNER_FAULT" in
                    search_exit) printf '[{"slug":"bat","name":"bat"}]\n'; exit 23 ;;
                    search_stderr) printf '[{"slug":"bat","name":"bat"}]\n' >&2; exit 23 ;;
                    search_stderr_only) printf '[{"slug":"bat","name":"bat"}]\n' >&2 ;;
                    search_multi) printf '[{"slug":"bat","name":"bat"}]\n[{"slug":"bat","name":"bat"}]\n' ;;
                    search_empty) printf '[]\n' ;;
                    search_wrong) printf '[{"slug":"batcat","name":"bat"}]\n' ;;
                    search_banner) printf "Found bat results\n" ;;
                    *) printf '[{"slug":"bat","name":"bat"}]\n' ;;
                esac ;;
            search:language=go)
                case "$RUNNER_FAULT" in
                    filter_exit) printf '[{"slug":"diffnav","language":"go"}]\n'; exit 23 ;;
                    filter_empty) printf '[]\n' ;;
                    filter_wrong) printf '[{"slug":"diffnav","language":"go"},{"slug":"bat","language":"rust"}]\n' ;;
                    *) printf '[{"slug":"diffnav","language":"go"}]\n' ;;
                esac ;;
            search:curl)
                case "$RUNNER_FAULT" in
                    user_exit) printf '[{"slug":"curl","name":"curl"}]\n'; exit 23 ;;
                    user_wrong) printf '[{"slug":"curlie","name":"curl"}]\n' ;;
                    *) printf '[{"slug":"curl","name":"curl"}]\n' ;;
                esac ;;
            install:btop)
                if [[ "$RUNNER_FAULT" == apk_display_wrong ]]; then
                    printf 'Error: apk add btop is unavailable\n'
                else
                    printf '│ linux:alpine │ apk add btop │\n'
                fi
                if [[ "$RUNNER_FAULT" == apk_display_exit ]]; then exit 23; fi ;;
            install:diffnav)
                if [[ "$RUNNER_FAULT" == go_display_wrong ]]; then
                    printf '│ go │ go install github.com/other/tool@latest │\n'
                else
                    printf '│ go │ go install github.com/dlvhdr/diffnav@latest │\n'
                fi
                if [[ "$RUNNER_FAULT" == go_display_exit ]]; then exit 23; fi ;;
            install:bat)
                if [[ "$*" == *--mise* ]]; then
                    case "$RUNNER_FAULT" in
                        mise_wrong) printf 'Error: cannot generate mise use command\n' ;;
                        *) printf 'mise use --global cargo:bat\n' ;;
                    esac
                    if [[ "$RUNNER_FAULT" == mise_exit ]]; then exit 23; fi
                else
                    printf 'Batch Install: 2 tools\n'
                    case "$RUNNER_FAULT" in
                        batch_banner) ;;
                        batch_failed) printf '✓ Completed: bat\n✗ Failed: btop - unavailable\n  ✓ Completed: 1\n  ✗ Failed: 1\n' ;;
                        batch_skipped) printf '✓ Completed: bat\n○ Skipped: btop\n  ✓ Completed: 1\n  ○ Skipped: 1\n' ;;
                        batch_wrong) printf '✓ Completed: bat\n✓ Completed: gomi\n  ✓ Completed: 2\n' ;;
                        batch_no_command) printf '✓ Completed: bat\n✓ Completed: btop\n  ✓ Completed: 2\n' ;;
                        *) printf 'Command: apk add bat\n✓ Completed: bat\nCommand: apk add btop\n✓ Completed: btop\n  ✓ Completed: 2\n' ;;
                    esac
                    if [[ "$RUNNER_FAULT" == batch_exit ]]; then exit 23; fi
                fi ;;
            *) exit 97 ;;
        esac ;;
    go)
        printf 'go version go1.25.6 linux/amd64\n'
        if [[ "$RUNNER_FAULT" == go_exit ]]; then exit 23; fi ;;
    sudo) exec "$@" ;;
    apk) if [[ "$RUNNER_FAULT" == apk_exec_exit ]]; then exit 23; fi ;;
    btop) exit 0 ;;
    su)
        [[ "$1" == "-" && "$2" == testuser && "$3" == "-c" ]] || exit 97
        exec bash -c "$4" ;;
    *) exit 97 ;;
esac
`

func runIntegrationFixture(t *testing.T, fault string) (string, error) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("integration runner regression checks require Bash")
	}
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("integration runner regression checks require jq")
	}
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.Mkdir(binDir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"troveler", "go", "sudo", "apk", "btop", "su"} {
		//nolint:gosec // G306: owner-only executables are required inside this test's private temporary directory.
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(fixtureCommand), 0700); err != nil {
			t.Fatal(err)
		}
	}
	resultsDir := filepath.Join(dir, "results")
	cmd := exec.CommandContext(t.Context(), "bash", "run_tests.sh")
	cmd.Env = append(os.Environ(), "PATH="+binDir+":"+os.Getenv("PATH"),
		"RESULTS_DIR="+resultsDir, "TMPDIR="+dir, "RUNNER_FAULT="+fault)
	output, err := cmd.CombinedOutput()
	//nolint:gosec // G304: this is the fixed results filename inside the test's private temporary directory.
	results, readErr := os.ReadFile(filepath.Join(resultsDir, "test_results.txt"))
	if readErr != nil {
		t.Fatalf("runner did not write results: %v\n%s", readErr, output)
	}

	return string(results), err
}

func TestIntegrationRunnerValidResults(t *testing.T) {
	results, err := runIntegrationFixture(t, "")
	if err != nil || strings.Count(results, "PASS:") != 9 || strings.Contains(results, "FAIL:") {
		t.Fatalf("valid commands rejected: %v\n%s", err, results)
	}
}

func TestIntegrationRunnerRejectsFalsePositives(t *testing.T) {
	cases := []struct {
		fault string
		check string
	}{
		{fault: "search_exit", check: "basic_search"},
		{fault: "search_stderr", check: "basic_search"},
		{fault: "search_stderr_only", check: "basic_search"},
		{fault: "search_multi", check: "basic_search"},
		{fault: "search_empty", check: "basic_search"},
		{fault: "search_wrong", check: "basic_search"},
		{fault: "search_banner", check: "basic_search"},
		{fault: "apk_display_exit", check: "apk_install_display"},
		{fault: "apk_display_wrong", check: "apk_install_display"},
		{fault: "apk_exec_exit", check: "apk_install_exec"},
		{fault: "go_display_exit", check: "go_install_display"},
		{fault: "go_display_wrong", check: "go_install_display"},
		{fault: "go_exit", check: "go_toolchain"},
		{fault: "mise_exit", check: "mise_integration"},
		{fault: "mise_wrong", check: "mise_integration"},
		{fault: "batch_exit", check: "batch_install"},
		{fault: "batch_banner", check: "batch_install"},
		{fault: "batch_failed", check: "batch_install"},
		{fault: "batch_skipped", check: "batch_install"},
		{fault: "batch_wrong", check: "batch_install"},
		{fault: "batch_no_command", check: "batch_install"},
		{fault: "filter_exit", check: "search_filters"},
		{fault: "filter_empty", check: "search_filters"},
		{fault: "filter_wrong", check: "search_filters"},
		{fault: "user_exit", check: "sudo_user_test"},
		{fault: "user_wrong", check: "sudo_user_test"},
	}
	for _, tc := range cases {
		t.Run(tc.fault, func(t *testing.T) {
			results, err := runIntegrationFixture(t, tc.fault)
			if err == nil {
				t.Fatalf("runner accepted %s\n%s", tc.fault, results)
			}
			if !strings.Contains(results, "FAIL: "+tc.check+" -") ||
				strings.Contains(results, "PASS: "+tc.check+" -") {
				t.Fatalf("wrong result for %s: %v\n%s", tc.check, err, results)
			}
			if strings.Count(results, "FAIL:") != 1 || strings.Count(results, "PASS:") != 8 {
				t.Fatalf("fault %s affected unrelated checks\n%s", tc.fault, results)
			}
		})
	}
}
