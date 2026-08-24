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
