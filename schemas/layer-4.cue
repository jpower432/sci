package schemas

#EvaluationResults: {
	"evaluation-set": [#ControlEvaluation, ...#ControlEvaluation] @go(EvaluationSet)
	...
}

#ControlEvaluation: {
	name:              string
	"control-id":      string @go(ControlId)
	result:            #Result
	message:           string
	"corrupted-state": bool @go(CorruptedState)
	assessments: [...#Assessment]
}

#Assessment: {
	"requirement-id": string @go(RequirementId)
	applicability: [...string]
	description: string
	result:      #Result
	message:     string
	// Procedures defines the assessment procedures associated with the assessment
	procedures: [...#AssessmentProcedure]
	"run-duration"?: string @go(RunDuration)
	value?:          _
	changes?: {[string]: #Change}
	recommendation?: string
}

// AssessmentProcedure describes a testing procedure for evaluating a Layer 2 control requirement.
#AssessmentProcedure: {
	// Id uniquely identifies the assessment procedure being executed
	id: string
	// Name provides a summary of the procedure
	name: string
	// Description provides a detailed explanation of the procedure
	description: string
	// Method describe the high-level method used to determine the results of the procedure
	method: #EvaluationMethod
	// Run is a boolean indicating whether the procedure was run or not. When run is true, result is expected to be present
	run: bool
	// Documentation provides a URL to documentation that describes how the assessment procedure evaluates the control requirement
	documentation?: #URL
	// Steps provides the address for the assessment steps executed
	steps: [...#AssessmentStep]
}

// Additional constraints on Assessment Procedure.
#AssessmentProcedure: {
	run: false
	// Message describes the result of the procedure
	message?: string
	// Result communicates the outcome(s) of the procedure
	result?: #Result
} | {
	run:      true
	message!: string
	result!:  #Result
}

#AssessmentStep: string

#Change: {
	"target-name":    string @go(TargetName)
	description:      string
	"target-object"?: _ @go(TargetObject)
	applied?:         bool
	reverted?:        bool
	error?:           string
}

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown"

// EvaluationMethod describes method options that can be used to determine the results
#EvaluationMethod: "Automated" | "Manual"

// URL describes a specific subset of URLs of interest to the framework
#URL: =~"^https?://[^\\s]+$"
