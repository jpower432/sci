package schemas

@go(layer3)

#PolicyDocument: {
	metadata: #Metadata
    contacts: #Contacts

    scope: #Scope

    guidance-references: [...#Mapping]
    control-references: [...#Mapping]
    control-modifications: [...#Modification]
}

#Metadata: {
	id: string
	title: string
    objective: string
	version: string
    contacts: #Contacts
	"last-modified": string @go(LastModified) @yaml("last-modified,omitempty")


	"organization-id"?: string @go(OrganizationID)

	author-notes?: string // For any instructions or considerations the author would like to provide for readers

	"mapping-references"?: [...#MappingReference] @go(MappingReferences)
}

#Contacts: {
    author: #Contact
    responsible: [...#Contact] // The person or group responsible for implementing controls for technical requirements
    accountable: [...#Contact] // The person or group accountable for evaluating and enforcing the efficacy of technical controls
    consulted?: [...#Contact] // Optional person or group who may be consulted for more information about the technical requirements
    informed?: [...#Contact] // Optional person or group who must recieve updates about compliance with this policy
}

#ImplementationPlan: {
    // The process through which notified parties should be made aware of this policy
    "notification-process"?: string @go(NotifactionProcess) @yaml("notification-process",omitempty)

    "notified-parties"?: [...#NotificationGroup] @go(NotifiedParties) @yaml("notified-parties",omitempty)

    evaluation: #ImplementationDetails
    "evaluation-points"?: [...#EvaluationPoint]

    enforcement: #ImplementationDetails
    "enforcement-methods"?: [...#EnforcementMethod] @go(EnforcementMethods) @yaml("enforcement-methods",omitempty)

    // The process that will be followed in the event that noncompliance is detected in an applicable resource
    "noncompliance-plan"?: string
}

#ImplementationDetails: {
    start: #Datetime
    end?: #Datetime
    notes: string
}

#EvaluationPoint: "development-tools" // For noncompliance risk to workflows or local machines
                | "pre-commit-hook" // For noncompliance risk to a development repository
                | "pre-merge" // For noncompliance risk to primary repositories
                | "pre-build" // For noncompliance risk to built assets
                | "pre-release" // For noncompliance risk to released assets
                | "pre-deploy" // For noncompliance risk to deployments
                | "runtime-adhoc" // For situations where drift may occur
                | "runtime-scheduled" // For situations where drift detection can be automated

#EnforcementMethod: "Deployment Gate"
                | "Autoremediation"
                | "Manual Remediation"

#NotificationGroup: "Responsible"
                | "Acccountable"
                | "Consulted"
                | "Informed"

#Mapping: {
	"reference-id": string @go(ReferenceId) @yaml("reference-id",omitempty)
    "in-scope": #Scope @go(InScope) @yaml("in-scope",omitempty)
    "out-of-scope": #Scope @go(OutOfScope) @yaml("out-of-scope",omitempty)
    control-modifications: [...#ControlModifier]
    assessment-requirement-modifications: [...#AssessmentRequirementModifier]
    guideline-modifications: [...#GuidelineModifier]
}

#Scope: {
    // geopolitical boundaries such as region names or jurisdictions
    boundaries?: [...string]

    // names of technology categories or services
    technologies?: [...string]

    // names of organizations who make the listed technologies available
    providers?: [...string]    
}

#ControlModifier: {
    target-id: string
    modification-type: #ModType
    rationale: string

    title?: string
    objective?: string
}

#AssessmentRequirementModifier: {
    target-id: string
    modification-type: #ModType
    rationale: string

    text: string
    applicability: [...string]

    recommendation?: string
}

#GuidelineModifier: {
    target-id: string
    modification-type: #ModType
    rationale: string

	title: string
	objective?: string
	recommendations?: [...string]
	"base-guideline-id"?: string @go(BaseGuidelineID) @yaml("base-guideline-id,omitempty")
	rationale?: #Rationale @go(Rationale,optional=nillable)
	"guideline-parts"?: [...#Part] @go(GuidelineParts) @yaml("guideline-parts,omitempty")
	"guideline-mappings"?: [...#Mapping] @go(GuidelineMappings) @yaml("guideline-mappings,omitempty")
	"principle-mappings"?: [...#Mapping] @go(PrincipleMappings) @yaml("principle-mappings,omitempty")
	"see-also"?: [...string] @go(SeeAlso) @yaml("see-also,omitempty")
	"external-references"?: [...string] @go(ExternalReferences) @yaml("external-references,omitempty")
}

#ModType: "tighten" | "clarify" | "loosen" | "exclude"
