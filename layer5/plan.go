package layer5

import (
	"github.com/google/uuid"

	"github.com/ossf/gemara/layer4"
)

// This function simulates the Enforcement Layer's logic.
// It takes a finding and makes a decision, generating a unique EnforcementID.

func EnforcementDecision(log layer4.AssessmentLog, target Target) EnforcementAction {
	enforcementID := uuid.New().String()
	action := Decision("Audit")

	// Simple logic: a failed evaluation leads to a "Mutate" decision.
	if log.Result == layer4.Failed {
		action = "Mutate"
	}

	return EnforcementAction{
		Id:       enforcementID,
		Decision: action,
		Target:   target,
		Finding: Finding{
			RequirementId: log.RequirementId,
			Result:        log.Result,
			Message:       log.Message,
		},
	}
}

// This function simulates the Remediation Layer's plan generation.
// It receives the EnforcementAction and embeds the EnforcementID.
func generateRemediationPlan(action EnforcementAction) RemediationPlan {
	planID := uuid.New().String()

	steps := []RemediationStep{
		{Id: "1", Name: "Enable Encryption"},
		{Id: "2", Name: "Update Access Policy"},
	}

	return RemediationPlan{
		Id:            planID,
		EnforcementId: action.Id,
		Steps:         steps,
	}
}
