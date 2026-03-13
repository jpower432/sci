// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// EnforcementLog records actions taken in response to noncompliance findings from Layer 5 evaluations.
#EnforcementLog: {
	// metadata provides detailed data about this log
	metadata: #Metadata @go(Metadata)
	// policy references the Policy being enforced
	policy: #ArtifactMapping @go(Policy)
	// target references the resource enforcement was performed on
	target: #Resource @go(Target)
	actions: [...#Action] @go(Actions,type=[]*Action)
}

// Action captures performed enforcement actions.
#Action: {
	// type is the enforcement action taken
	type: #ActionType @go(Type)

	// method identifies the Policy enforcement method
	method: #MethodType @go(Method)

	// status is the outcome of the enforcement action
	status: #Status @go(Status)

	// message provides additional context about the action
	message?: string @go(Message,type=*string)

	// executed-at is the timestamp when the action was taken
	"executed-at": #Datetime @go(ExecutedAt)

	// procedures references the code paths or addresses that carried out this enforcement action
	procedures: [...#Procedure]

	// justification links the action to its assessment findings and any applicable exceptions
	justification?: #Justification @go(Justification,optional=nillable)
}

// Procedure is a reference to the code that performed an enforcement action
#Procedure: string @go(-)

// Justification provides the assessment data and exception references that justify an enforcement action.
#Justification: {
	// assessments links the action to one or more Assessment Findings
	assessments: [...#AssessmentFinding] @go(Assessments)

	// exceptions references approved Policy exceptions that authorize the action
	exceptions?: [...#ArtifactMapping] @go(Exceptions)
}

// AssessmentFinding maps an enforcement action to its originating assessment data across Layer 2, Layer 3, and Layer 5.
#AssessmentFinding: {
	// requirement maps to the Layer 2 assessment requirement that was evaluated
	requirement?: #EntryMapping @go(Requirement)

	// plan maps to the Policy assessment plan that was executed
	plan?: #EntryMapping @go(Plan)

	// log maps to the EvaluationLog entry containing the finding
	log: #EntryMapping @go(Log)
}

// ActionType enumerates the possible enforcement actions.
#ActionType: "Block" | "Allow" | "Remediate" | "Waive" @go(-)

// Status enumerates the possible outcomes of an enforcement action.
#Status: "Success" | "Failure" | "Not Run" @go(-)
