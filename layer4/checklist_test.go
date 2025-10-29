package layer4

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToChecklist(t *testing.T) {
	tests := []struct {
		testName     string
		plan         EvaluationPlan
		expectedFile string
	}{
		{
			testName: "Full plan",
			plan: EvaluationPlan{
				Metadata: Metadata{
					Id:      "test-plan-001",
					Version: "1.0.0",
					Author: Author{
						Name:    "Test Author",
						Uri:     "https://example.com",
						Version: "1.0.0",
					},
					MappingReferences: []MappingReference{
						{
							Id:          "ref-1",
							Title:       "Test Framework",
							Version:     "v1.0",
							Description: "Test framework description",
							Url:         "https://example.com/framework",
						},
					},
				},
				Plans: []AssessmentPlan{
					{
						Control: Mapping{
							ReferenceId: "REF-1",
							EntryId:     "CTRL-1",
							Remarks:     "This is a test control",
						},
						Assessments: []Assessment{
							{
								Requirement: Mapping{
									ReferenceId: "REF-1",
									EntryId:     "REQ-1",
									Remarks:     "This is a test requirement",
								},
								Procedures: []AssessmentProcedure{
									{
										Id:            "PROC-1",
										Name:          "Test Procedure 1",
										Description:   "This is a test procedure description",
										Documentation: "https://example.com/doc/proc1",
									},
									{
										Id:          "PROC-2",
										Name:        "Test Procedure 2",
										Description: "Another test procedure\nwith multiple lines",
									},
								},
							},
						},
					},
				},
			},
			expectedFile: "checklist_full_plan.golden.md",
		},
		{
			testName: "Control with unknown entry ID",
			plan: EvaluationPlan{
				Metadata: Metadata{
					Id: "unknown-control-id",
					Author: Author{
						Name: "Test Author",
					},
				},
				Plans: []AssessmentPlan{
					{
						Control: Mapping{
							EntryId: "",
						},
						Assessments: []Assessment{
							{
								Requirement: Mapping{
									EntryId: "REQ-1",
								},
								Procedures: []AssessmentProcedure{
									{
										Id:   "PROC-1",
										Name: "Procedure 1",
									},
								},
							},
						},
					},
				},
			},
			expectedFile: "checklist_unknown_control_id.golden.md",
		},
		{
			testName: "Requirement with unknown entry ID",
			plan: EvaluationPlan{
				Metadata: Metadata{
					Id: "unknown-requirement-id",
					Author: Author{
						Name: "Test Author",
					},
				},
				Plans: []AssessmentPlan{
					{
						Control: Mapping{
							EntryId: "CTRL-1",
						},
						Assessments: []Assessment{
							{
								Requirement: Mapping{
									EntryId: "",
								},
								Procedures: []AssessmentProcedure{
									{
										Id:   "PROC-1",
										Name: "Procedure 1",
									},
								},
							},
						},
					},
				},
			},
			expectedFile: "checklist_unknown_requirement_id.golden.md",
		},
		{
			testName: "Procedure with ID only (no name)",
			plan: EvaluationPlan{
				Metadata: Metadata{
					Id: "id-only-procedure",
					Author: Author{
						Name: "Test Author",
					},
				},
				Plans: []AssessmentPlan{
					{
						Control: Mapping{
							EntryId: "CTRL-1",
						},
						Assessments: []Assessment{
							{
								Requirement: Mapping{
									EntryId: "REQ-1",
								},
								Procedures: []AssessmentProcedure{
									{
										Id: "PROC-1",
									},
								},
							},
						},
					},
				},
			},
			expectedFile: "checklist_id_only_procedure.golden.md",
		},
		{
			testName: "Procedure without ID or name",
			plan: EvaluationPlan{
				Metadata: Metadata{
					Id: "no-id-name-procedure",
					Author: Author{
						Name: "Test Author",
					},
				},
				Plans: []AssessmentPlan{
					{
						Control: Mapping{
							EntryId: "CTRL-1",
						},
						Assessments: []Assessment{
							{
								Requirement: Mapping{
									EntryId: "REQ-1",
								},
								Procedures: []AssessmentProcedure{
									{
										Description: "Procedure with description only",
									},
								},
							},
						},
					},
				},
			},
			expectedFile: "checklist_no_id_name_procedure.golden.md",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			checklist, err := test.plan.ToChecklist()
			require.NoError(t, err)

			expected := readExpectedFile(t, test.expectedFile)
			assert.Equal(t, expected, checklist)
		})
	}
}

func readExpectedFile(t *testing.T, filename string) string {
	t.Helper()
	expectedPath := filepath.Join("test-data", filename)

	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	return string(content)
}