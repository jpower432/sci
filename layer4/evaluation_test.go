package layer4

import (
	"testing"

	"github.com/revanite-io/sci/layer2"
	"github.com/stretchr/testify/assert"
)

func TestNewEvaluationWithEmptyCatalog(t *testing.T) {
	catalog := layer2.Catalog{
		Metadata: &layer2.Metadata{
			Id: "SomeControlsCatalog",
		},
	}
	evaluation := NewEvaluation(catalog)

	assert.Equalf(t, catalog.Metadata.Id, evaluation.CatalogID,
		"expected CatalogID to be '%s', got '%s'", catalog.Metadata.Id, evaluation.CatalogID)
	assert.Equal(t, len(catalog.ControlFamilies), len(evaluation.ControlEvaluations),
		"expected Evaluations to be %d, got %d", len(catalog.ControlFamilies), len(evaluation.ControlEvaluations))
}

func TestNewEvaluationWithBasicCatalog(t *testing.T) {
	catalog := layer2.Catalog{
		Metadata: &layer2.Metadata{
			Id: "SomeControlsCatalog",
		},
		ControlFamilies: []layer2.ControlFamily{
			{
				Title:       "Control Family 1",
				Description: "Description of Control Family 1",
				Controls: []layer2.Control{
					{
						Id:        "Control1",
						Title:     "Control 1",
						Objective: "Objective of Control 1",
						AssessmentRequirements: []layer2.AssessmentRequirement{
							{
								Id:   "Requirement1",
								Text: "Text of Requirement 1",
								Applicability: []string{
									"applicable",
								},
							},
							{
								Id:   "Requirement2",
								Text: "Text of Requirement 2",
								Applicability: []string{
									"applicable",
								},
							},
						},
					},
				},
			},
			{
				Title:       "Control Family 2",
				Description: "Description of Control Family 2",
				Controls: []layer2.Control{
					{
						Id:        "Control2",
						Title:     "Control 2",
						Objective: "Objective of Control 2",
						AssessmentRequirements: []layer2.AssessmentRequirement{
							{
								Id:   "Requirement1",
								Text: "Text of Requirement 1",
								Applicability: []string{
									"applicable",
								},
							},
						},
					},
				},
			},
		},
	}
	evaluation := NewEvaluation(catalog)

	assert.Equalf(t, catalog.Metadata.Id, evaluation.CatalogID,
		"expected FrameworkID to be '%s', got '%s'", catalog.Metadata.Id, evaluation.CatalogID)

	// the evaluation should contain 1 ControlEvaluation for every control in the provided catalog
	controlEvals := evaluation.ControlEvaluations
	assert.Equal(t, 2, len(controlEvals),
		"expected Evaluations to contain %d items, got %d", len(catalog.ControlFamilies[0].Controls), len(controlEvals))

	// each ControlEvaluation should have a reference to the control name and ID
	assert.Equal(t, catalog.ControlFamilies[0].Controls[0].Id, controlEvals[0].ControlID,
		"expected ControlID to be '%s', got '%s'", catalog.ControlFamilies[0].Controls[0].Id, controlEvals[0].ControlID)
	assert.Equal(t, catalog.ControlFamilies[1].Controls[0].Id, controlEvals[1].ControlID,
		"expected ControlID to be '%s', got '%s'", catalog.ControlFamilies[1].Controls[0].Id, controlEvals[1].ControlID)

	// each ControlEvaluation should have 1 Assessment for each requirement in the control
	// first
	assert.Equal(t, len(catalog.ControlFamilies[0].Controls[0].AssessmentRequirements), len(controlEvals[0].Assessments),
		"expected Control1 to have %d assessments, got %d", len(catalog.ControlFamilies[0].Controls[0].AssessmentRequirements), len(controlEvals[0].Assessments))
	assert.Equal(t, catalog.ControlFamilies[0].Controls[0].AssessmentRequirements[0].Id, controlEvals[0].Assessments[0].RequirementID,
		"expected RequirementID to be '%s', got '%s'", catalog.ControlFamilies[0].Controls[0].AssessmentRequirements[0].Id, controlEvals[0].Assessments[0].RequirementID)
	assert.Equal(t, catalog.ControlFamilies[0].Controls[0].AssessmentRequirements[1].Id, controlEvals[0].Assessments[1].RequirementID,
		"expected RequirementID to be '%s', got '%s'", catalog.ControlFamilies[0].Controls[0].AssessmentRequirements[1].Id, controlEvals[0].Assessments[1].RequirementID)
	// second
	assert.Equal(t, len(catalog.ControlFamilies[1].Controls[0].AssessmentRequirements), len(controlEvals[1].Assessments),
		"expected Control2 to have %d assessments, got %d", len(catalog.ControlFamilies[1].Controls[0].AssessmentRequirements), len(controlEvals[1].Assessments))
	assert.Equal(t, catalog.ControlFamilies[1].Controls[0].AssessmentRequirements[0].Id, controlEvals[1].Assessments[0].RequirementID,
		"expected RequirementID to be '%s', got '%s'", catalog.ControlFamilies[0].Controls[0].AssessmentRequirements[0].Id, controlEvals[1].Assessments[0].RequirementID)

	// every Assessment should have 0 methods, because these cannot be inferred from the layer 2 catalog
	assert.Equal(t, 0, len(controlEvals[0].Assessments[0].Methods), "expected Control1 Requirement1 to have 0 methods, got %d", len(controlEvals[0].Assessments[0].Methods))
	assert.Equal(t, 0, len(controlEvals[0].Assessments[1].Methods), "expected Control1 Requirement2 to have 0 methods, got %d", len(controlEvals[0].Assessments[1].Methods))
	assert.Equal(t, 0, len(controlEvals[1].Assessments[0].Methods), "expected Control2 Requirement1 to have 0 methods, got %d", len(controlEvals[1].Assessments[0].Methods))
}
