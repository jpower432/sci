package schemas

import "time"

#EvaluationResults: {
	"evaluation-set": [#ControlEvaluation, ...#ControlEvaluation] @go(EvaluationSet)
	...
}

#EvaluationPlan: {
	author: #Contact
	plans: [...#AssessmentPlan]
	...
}

#AssessmentPlan: {
	"control-id": string @go(ControlId)
	"assessment": [...#Assessment] @go(Assessment)
}

#Assessment: {
	"requirement-id": string @go(RequirementId)
	procedures: [...#AssessmentProcedure] @go(Procedures)
}

#ControlEvaluation: {
	name:              string
	"control-id":      string @go(ControlId)
	result:            #Result
	"corrupted-state": bool @go(CorruptedState)
	"assessments-logs": [...#AssessmentLog] @go(AssessmentLogs)
}

#AssessmentLog: {
	"requirement-id": string @go(RequirementId)
	"procedure-id"?:  string @go(ProcedureId)
	applicability: [...string]
	description: string
	result:      #Result
	steps: [...#AssessmentStep]
	"steps-executed"?: int @go(StepsExecuted)
	"start":           #Datetime
	"end"?:            #Datetime
	value?:            _
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
	// Documentation provides a URL to documentation that describes how the assessment procedure evaluates the control requirement
	documentation?: =~"^https?://[^\\s]+$"
}

#AssessmentStep: string

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

#Contact: {
	// The contact person's name.
	name: string
	// Indicates whether this admin is the first point of contact for inquiries. Only one entry should be marked as primary.
	primary: bool
	// The entity with which the contact is affiliated, such as a school or employer.
	affiliation?: string @go(Affiliation,type=*string)
	// A preferred email address to reach the contact.
	email?: #Email @go(Email,type=*Email)
	// A social media handle or profile for the contact.
	social?: string @go(Social,type=*string)
}
#Email: =~"^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$"

#Result: {
	status:  #Status
	message: string
}

#Status: "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown"
