package layer4

import "github.com/ossf/gemara/schemas/common"

@go(layer4)

// EvaluationLog contains the results of evaluating a set of Layer 2 controls.
#EvaluationLog: {
	"metadata"?: common.#Metadata @go(Metadata)
	"evaluations": [#ControlEvaluation, ...#ControlEvaluation] @go(Evaluations,type=[]*ControlEvaluation)
}

// ControlEvaluation contains the results of evaluating a single Layer 4 control.
#ControlEvaluation: {
	name:    string
	result:  #Result
	message: string
	control: common.#SimpleMapping
	// Enforce that control reference and the assessments' references match
	// This formulation uses the control's reference if the assessment doesn't include a reference
	"assessment-logs": [...#AssessmentLog & {
		if control."reference-id" != _|_ {
			requirement: {
				"reference-id": control."reference-id"
			}
		}
	}] @go(AssessmentLogs,type=[]*AssessmentLog)
}

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: common.#SimpleMapping
	// Procedure should map to the assessment procedure being executed.
	description: string
	// Result is the overall outcome of the assessment procedure, matching the result of the last step that was run.
	result: #Result
	// Message provides additional context about the assessment result.
	message: string
	// Applicability is elevated from the Layer 2 Assessment Requirement to aid in execution and reporting.
	applicability: [...string] @go(Applicability,type=[]string)
	// Steps are sequential actions taken as part of the assessment, which may halt the assessment if a failure occurs.
	steps: [...#AssessmentStep]
	// Steps-executed is the number of steps that were executed as part of the assessment.
	"steps-executed"?: int @go(StepsExecuted)
	// Start is the timestamp when the assessment began.
	start: common.#Datetime
	// End is the timestamp when the assessment concluded.
	end?: common.#Datetime
	// Recommendation provides guidance on how to address a failed assessment.
	recommendation?: string
}

#AssessmentStep: string @go(-)

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
