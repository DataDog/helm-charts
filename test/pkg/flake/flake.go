// Package flake provides helpers to mark known-flaky tests.
//
// When the environment variable GO_TEST_SKIP_FLAKE is set to "true", any test
// marked with flake.Mark will be skipped instead of run. On main/release
// pipelines where GO_TEST_SKIP_FLAKE is not set (or set to "false"), the test
// runs normally; a sentinel log line is emitted so that CI post-processors can
// distinguish known-flaky failures from real ones.
//
// Ported from github.com/DataDog/datadog-agent/pkg/util/testutil/flake.
package flake

import (
	"os"
	"testing"
)

// KnownFlakyMessage is the sentinel string logged when a flaky test runs.
// CI post-processors match this string to allow failures on marked tests.
const KnownFlakyMessage = "flakytest: this is a known flaky test"

// Mark marks t as a known-flaky test.
//
//   - If GO_TEST_SKIP_FLAKE=true the test is skipped immediately.
//   - Otherwise the test runs as usual and KnownFlakyMessage is logged so that
//     CI tooling can identify and allow the failure.
func Mark(t *testing.T) {
	t.Helper()
	if os.Getenv("GO_TEST_SKIP_FLAKE") == "true" {
		t.Skip(KnownFlakyMessage)
	} else {
		t.Log(KnownFlakyMessage)
	}
}
