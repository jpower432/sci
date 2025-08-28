package schemas

import "time"

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
	procedures: [...#AssessmentProcedure]
	"start":           #Datetime
	"end"?:            #Datetime
	value?:            _
	changes?: {[string]: #Change}
	recommendation?: string

// AssessmentProcedure describes a testing procedure for evaluating a Layer 2 control requirement.
#AssessmentProcedure: {
	// Id uniquely identifies the assessment procedure being executed
	id: string
	// Name provides a summary of the procedure
	name: string
	// Description provides a detailed explanation of the procedure
	description: string
	// Method describe the high-level method used to determine the results of the procedure
	method: #ProcedureMethod
	// Run is a boolean indicating whether the procedure was run or not. When run is true, result is expected to be present
	run: bool
	// Documentation provides a URL to documentation that describes how the assessment procedure evaluates the control requirement
	documentation?: #URL
	// Steps provides the address for the assessment steps executed
	"steps"?: [...#AssessmentStep]
}

// Additional constraints on Assessment Procedure.
#AssessmentProcedure: {
	run: false
	// Message describes the result of the procedure
	message?: string
	// Result communicates the outcome(s) of the procedure
	result?: ("Not Run" | *null) @go(Result,optional=nillable)
} | {
	run:     true
	message!: string
	result!: #ResultWhenRun
}

// Result describes valid assessment outcomes before and after execution.
#Result: #ResultWhenRun | "Not Run"
#AssessmentStep: string

// Result describes the outcome(s) of an assessment procedure when it is executed.
#ResultWhenRun: "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown"

// ProcedureMethod describes method options that can be used to determine the results
#ProcedureMethod: "Test" | "Observation"

// Change is a struct that contains the data and functions associated with a single change to a target resource.
#Change: {
	"target-name":    string @go(TargetName)
	description:      string
	"target-object"?: _ @go(TargetObject)
	applied?:         bool
	reverted?:        bool
	error?:           string
}

#Datetime: time.Format("2006-01-02T15:04:05Z07:00") @go(Datetime,format="date-time")
