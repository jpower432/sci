// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBreakingCheckRequiresBase(t *testing.T) {
	// No basePath supplied: the check must error rather than silently pass.
	if _, err := runBreakingCheck(breakingCheckOpts{candidatePath: "testdata/base.yaml"}); err == nil {
		t.Fatalf("expected error when --base is missing, got nil")
	}
}

// TestRunBreakingCheckAllowSuppresses exercises the full --allow path: a real
// detected breaking change is suppressed when its "<id> <schema>" entry is in
// the allowlist, and the check then passes.
func TestRunBreakingCheckAllowSuppresses(t *testing.T) {
	// Without an allowlist, dropping a required property is breaking (exit 1).
	code, err := runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/rev_remove.yaml",
	})
	if err != nil {
		t.Fatalf("runBreakingCheck: %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1 without allowlist, got %d", code)
	}

	// Allowlisting that exact "<id> <schema>" pair suppresses it (exit 0).
	allow := filepath.Join(t.TempDir(), ".oasdiff-allow")
	content := "# result became optional is an accepted, reviewed change\nresponse-property-became-optional ControlEvaluation\n"
	if err := os.WriteFile(allow, []byte(content), 0644); err != nil {
		t.Fatalf("writing allowlist: %v", err)
	}
	code, err = runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/rev_remove.yaml",
		allowPath:     allow,
	})
	if err != nil {
		t.Fatalf("runBreakingCheck (allowed): %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 with matching allowlist entry, got %d", code)
	}
}

func TestRunBreakingCheckLocalBaseline(t *testing.T) {
	// base vs a mutated schema, both provided as local OpenAPI files, exercises
	// the diff+exit-code path without touching the network or CUE generation.
	code, err := runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/rev_remove.yaml",
	})
	if err != nil {
		t.Fatalf("runBreakingCheck: %v", err)
	}
	if code != 1 {
		t.Fatalf("expected exit code 1 for a breaking change, got %d", code)
	}
	code, err = runBreakingCheck(breakingCheckOpts{
		basePath:      "testdata/base.yaml",
		candidatePath: "testdata/base.yaml",
	})
	if err != nil {
		t.Fatalf("runBreakingCheck: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 for no change, got %d", code)
	}
}
