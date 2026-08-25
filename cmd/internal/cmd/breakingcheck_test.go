// SPDX-License-Identifier: Apache-2.0

package cmd

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

func TestLoadAllowlistOpenErrorPropagates(t *testing.T) {
	// A directory opens but cannot be read as a file; the scan/read (or on some
	// platforms the open) must surface a non-nil error rather than being silenced.
	dir := t.TempDir()
	if _, err := loadAllowlist(dir); err == nil {
		t.Fatalf("expected error for directory path, got nil")
	}

	// A non-existent path must also error rather than returning a nil map silently.
	missing := filepath.Join(dir, "does-not-exist.txt")
	got, err := loadAllowlist(missing)
	if err == nil {
		t.Fatalf("expected error for missing path, got nil")
	}
	if got != nil {
		t.Fatalf("expected nil map on error, got %+v", got)
	}
}

func TestLoadAllowlistEmptyPath(t *testing.T) {
	got, err := loadAllowlist("")
	if err != nil {
		t.Fatalf("loadAllowlist(\"\"): %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil empty map for empty path")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %+v", got)
	}
}

func TestLoadAllowlistParsesEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allow.txt")
	content := "# a comment\n\nfirst-id\n  second-id  \n# another\nthird-id\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing allowlist: %v", err)
	}
	got, err := loadAllowlist(path)
	if err != nil {
		t.Fatalf("loadAllowlist: %v", err)
	}
	for _, id := range []string{"first-id", "second-id", "third-id"} {
		if !got[id] {
			t.Fatalf("expected %q in allowlist, got %+v", id, got)
		}
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %+v", got)
	}
}

func TestFilterAllowed(t *testing.T) {
	changes := []Change{
		{ID: "a", Path: "/_schema/Foo", Text: "x"},
		{ID: "b", Path: "/_schema/Foo", Text: "y"},
	}
	got := filterAllowed(changes, map[string]bool{"a /_schema/Foo": true})
	if len(got) != 1 || got[0].ID != "b" {
		t.Fatalf("expected only [b], got %+v", got)
	}
}

func TestFilterAllowedPathScoped(t *testing.T) {
	changes := []Change{
		{ID: "a", Path: "/_schema/Foo", Text: "x"},
		{ID: "a", Path: "/_schema/Bar", Text: "y"},
	}
	// A path-scoped entry drops only the matching change, not every change
	// sharing the check ID.
	got := filterAllowed(changes, map[string]bool{"a /_schema/Foo": true})
	if len(got) != 1 || got[0].Path != "/_schema/Bar" {
		t.Fatalf("expected only the Bar change to remain, got %+v", got)
	}
	// A bare check ID (no path) matches nothing: entries must be path-scoped.
	got = filterAllowed(changes, map[string]bool{"a": true})
	if len(got) != 2 {
		t.Fatalf("expected a bare ID to drop nothing, got %+v", got)
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
