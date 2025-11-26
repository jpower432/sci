package layer3

import (
	"testing"

	"github.com/ossf/gemara/common"
	"github.com/ossf/gemara/layer2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToChecklist(t *testing.T) {
	tests := []struct {
		name     string
		policy   *Policy
		wantErr  bool
		validate func(*testing.T, Checklist)
	}{
		{
			name: "Policy with control references",
			policy: &Policy{
				Metadata: common.Metadata{
					Id:      "test-policy-001",
					Version: "1.0.0",
					Author: common.Actor{
						Id:   "author-001",
						Name: "Test Author",
						Type: common.Human,
					},
					MappingReferences: []common.MappingReference{
						{
							Id:    "OSPS-B",
							Title: "Open Source Project Security Baseline",
							Url:   "file://../layer2/test-data/good-osps.yml",
						},
					},
				},
				ControlReferences: []PolicyMapping{
					{
						ReferenceId: "OSPS-B",
						ControlModifications: []ControlModifier{
							{
								TargetId: "OSPS-AC-01",
								Overrides: &layer2.Control{
									Title: "Enhanced Multi-Factor Authentication",
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, checklist Checklist) {
				assert.Equal(t, "test-policy-001", checklist.PolicyId)
				assert.Equal(t, "Test Author", checklist.Author)
				require.Len(t, checklist.Sections, 1)
				assert.Equal(t, "OSPS-AC-01", checklist.Sections[0].ControlName)
				// Should have items from the loaded catalog
				assert.Greater(t, len(checklist.Sections[0].Items), 0, "should have checklist items from catalog")
			},
		},
		{
			name: "Policy with multiple control references",
			policy: &Policy{
				Metadata: common.Metadata{
					Id:      "test-policy-002",
					Version: "1.0.0",
					Author: common.Actor{
						Id:   "author-001",
						Name: "Test Author",
						Type: common.Human,
					},
					MappingReferences: []common.MappingReference{
						{
							Id:    "OSPS-B",
							Title: "Open Source Project Security Baseline",
							Url:   "file://../layer2/test-data/good-osps.yml",
						},
					},
				},
				ControlReferences: []PolicyMapping{
					{
						ReferenceId: "OSPS-B",
						ControlModifications: []ControlModifier{
							{
								TargetId: "OSPS-AC-01",
							},
						},
					},
					{
						ReferenceId: "OSPS-B",
						ControlModifications: []ControlModifier{
							{
								TargetId: "OSPS-BR-01",
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, checklist Checklist) {
				require.Len(t, checklist.Sections, 2)
				assert.Equal(t, "OSPS-AC-01", checklist.Sections[0].ControlName)
				assert.Equal(t, "OSPS-BR-01", checklist.Sections[1].ControlName)
				// Should have items from the loaded catalog
				assert.Greater(t, len(checklist.Sections[0].Items), 0, "should have checklist items from catalog")
				assert.Greater(t, len(checklist.Sections[1].Items), 0, "should have checklist items from catalog")
			},
		},
		{
			name: "Policy with empty control references",
			policy: &Policy{
				Metadata: common.Metadata{
					Id:      "test-policy-003",
					Version: "1.0.0",
					Author: common.Actor{
						Id:   "author-001",
						Name: "Test Author",
						Type: common.Human,
					},
				},
				ControlReferences: []PolicyMapping{},
			},
			wantErr: false,
			validate: func(t *testing.T, checklist Checklist) {
				assert.Equal(t, "test-policy-003", checklist.PolicyId)
				assert.Empty(t, checklist.Sections)
			},
		},
		{
			name: "Policy with control reference without modifications",
			policy: &Policy{
				Metadata: common.Metadata{
					Id:      "test-policy-004",
					Version: "1.0.0",
					Author: common.Actor{
						Id:   "author-001",
						Name: "Test Author",
						Type: common.Human,
					},
					MappingReferences: []common.MappingReference{
						{
							Id:    "OSPS-B",
							Title: "Open Source Project Security Baseline",
							Url:   "file://../layer2/test-data/good-osps.yml",
						},
					},
				},
				ControlReferences: []PolicyMapping{
					{
						ReferenceId: "OSPS-B",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, checklist Checklist) {
				assert.Empty(t, checklist.Sections, "should have no sections when no control modifications")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checklist, err := tt.policy.ToChecklist()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, checklist)
			}
		})
	}
}

func Test_ToMarkdownChecklist(t *testing.T) {
	tests := []struct {
		name     string
		policy   *Policy
		wantErr  bool
		validate func(*testing.T, string)
	}{
		{
			name: "Policy with control references generates markdown",
			policy: &Policy{
				Metadata: common.Metadata{
					Id:      "test-policy-001",
					Version: "1.0.0",
					Author: common.Actor{
						Id:   "author-001",
						Name: "Test Author",
						Type: common.Human,
					},
					MappingReferences: []common.MappingReference{
						{
							Id:    "OSPS-B",
							Title: "Open Source Project Security Baseline",
							Url:   "file://../layer2/test-data/good-osps.yml",
						},
					},
				},
				ControlReferences: []PolicyMapping{
					{
						ReferenceId: "OSPS-B",
						ControlModifications: []ControlModifier{
							{
								TargetId: "OSPS-AC-01",
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, markdown string) {
				assert.Contains(t, markdown, "test-policy-001")
				assert.Contains(t, markdown, "Test Author")
				assert.Contains(t, markdown, "OSPS-AC-01")
			},
		},
		{
			name: "Empty policy generates markdown",
			policy: &Policy{
				Metadata: common.Metadata{
					Id:      "test-policy-002",
					Version: "1.0.0",
					Author: common.Actor{
						Id:   "author-001",
						Name: "Test Author",
						Type: common.Human,
					},
				},
				ControlReferences: []PolicyMapping{},
			},
			wantErr: false,
			validate: func(t *testing.T, markdown string) {
				assert.Contains(t, markdown, "test-policy-002")
				assert.Contains(t, markdown, "Test Author")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, err := tt.policy.ToMarkdownChecklist()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, markdown)
			}
		})
	}
}
