// SPDX-License-Identifier: Apache-2.0

package cmd

import "testing"

func hasID(changes []Change, id string) bool {
	for _, c := range changes {
		if c.ID == id {
			return true
		}
	}
	return false
}

func TestBreakingChanges(t *testing.T) {
	base, err := loadWrapped("testdata/base.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	cases := []struct {
		name   string
		rev    string
		wantID string // "" means expect zero ERR changes
	}{
		{"identical", "testdata/base.yaml", ""},
		{"removed_required", "testdata/rev_remove.yaml", "response-property-became-optional"},
		{"added_required", "testdata/rev_add.yaml", "new-required-request-property"},
		{"enum_narrowed", "testdata/rev_enum_narrow.yaml", "request-property-enum-value-removed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rev, err := loadWrapped(tc.rev)
			if err != nil {
				t.Fatalf("load rev: %v", err)
			}
			got, err := breakingChanges(base, rev)
			if err != nil {
				t.Fatalf("breakingChanges: %v", err)
			}
			if tc.wantID == "" {
				if len(got) != 0 {
					t.Fatalf("expected no breaking changes, got %d: %+v", len(got), got)
				}
				return
			}
			if !hasID(got, tc.wantID) {
				t.Fatalf("expected change ID %q, got %+v", tc.wantID, got)
			}
		})
	}
}

// TestBreakingChangesExemptsExperimental confirms that a schema carrying
// x-status "experimental" is exempted from the gate: loadWrapped maps it to
// oasdiff's x-stability-level "alpha", which the default (beta) threshold
// filters out before the breaking-change checks run. A break that would fail
// for a stable schema must produce zero ERR changes here.
func TestBreakingChangesExemptsExperimental(t *testing.T) {
	base, err := loadWrapped("testdata/base_experimental.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	rev, err := loadWrapped("testdata/rev_experimental_break.yaml")
	if err != nil {
		t.Fatalf("load rev: %v", err)
	}
	got, err := breakingChanges(base, rev)
	if err != nil {
		t.Fatalf("breakingChanges: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected experimental schema to be exempt, got %d changes: %+v", len(got), got)
	}
}

// TestBreakingChangesCarriesSchema confirms each reported change carries the
// schema name (the internal "/_schema/" wrapper stripped) so multi-schema
// output is attributable and the allowlist can be keyed per schema.
func TestBreakingChangesCarriesSchema(t *testing.T) {
	base, err := loadWrapped("testdata/base.yaml")
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	rev, err := loadWrapped("testdata/rev_remove.yaml")
	if err != nil {
		t.Fatalf("load rev: %v", err)
	}
	got, err := breakingChanges(base, rev)
	if err != nil {
		t.Fatalf("breakingChanges: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected at least one breaking change")
	}
	for _, c := range got {
		if c.Schema != "ControlEvaluation" {
			t.Fatalf("expected schema %q, got %+v", "ControlEvaluation", c)
		}
	}
}
