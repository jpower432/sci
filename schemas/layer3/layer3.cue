package layer3

import (
	"github.com/ossf/gemara/schemas/common"
	"github.com/ossf/gemara/schemas/layer2"
	"github.com/ossf/gemara/schemas/layer1"
)

@go(layer3)

// Core Document Structure
#Policy: {
	metadata:          common.#Metadata
	"organization-id": string @go(OrganizationID) @yaml("organization-id")
	title:             string
	purpose:           string
	contacts:          #Contacts
	scope:             #Scope
	"imported-policies"?: [...#ImportedPolicy] @go(ImportedPolicies) @yaml("imported-policies,omitempty")
	"implementation-plan"?: #ImplementationPlan @go(ImplementationPlan) @yaml("implementation-plan,omitempty")
	"guidance-references"?: [...#PolicyMapping] @go(GuidanceReferences) @yaml("guidance-references")
	"control-references"?: [...#PolicyMapping] @go(ControlReferences) @yaml("control-references")
}

#Contacts: {
	responsible: [...common.#Contact] // The person or group responsible for implementing controls for technical requirements
	accountable: [...common.#Contact] // The person or group accountable for evaluating and enforcing the efficacy of technical controls
	consulted?: [...common.#Contact] // Optional person or group who may be consulted for more information about the technical requirements
	informed?: [...common.#Contact] // Optional person or group who must receive updates about compliance with this policy
}

#ImplementationPlan: {
	// The process through which notified parties should be made aware of this policy
	"notification-process"?: string @go(NotificationProcess) @yaml("notification-process,omitempty")
	"notified-parties"?: [...#NotificationGroup] @go(NotifiedParties) @yaml("notified-parties,omitempty")

	"evaluation-timeline": #ImplementationDetails @go(EvaluationTimeline) @yaml("evaluation-timeline")
	evaluators?: [...common.#Actor] @go(Evaluators) @yaml("evaluators,omitempty")

	"enforcement-timeline": #ImplementationDetails @go(EnforcementTimeline) @yaml("enforcement-timeline")
	"enforcement-methods"?: [...#EnforcementMethod] @go(EnforcementMethods) @yaml("enforcement-methods,omitempty")

	// The consequence that will be applied in the event that noncompliance is detected
	"noncompliance-consequence"?: string @go(NoncomplianceConsequence) @yaml("noncompliance-consequence,omitempty")
}

#ImplementationDetails: {
	start: common.#Datetime
	end?:  common.#Datetime
	notes: string
}

#Scope: {
	// geopolitical boundaries such as region names or jurisdictions
	boundaries?: [...string]
	// names of technology categories or services
	technologies?: [...string]
	// names of organizations who make the listed technologies available
	providers?: [...string]
}

// Layer 3 specific mapping that extends common Mapping with modifications
#PolicyMapping: {
	"reference-id": string @go(ReferenceId) @yaml("reference-id")
	"control-modifications"?: [...#ControlModifier] @go(ControlModifications) @yaml("control-modifications,omitempty")
	"guideline-modifications"?: [...#GuidelineModifier] @go(GuidelineModifications) @yaml("guideline-modifications,omitempty")
}

// Modifier Types
#ControlModifier: {
	id?:                      string   @go(Id)
	"target-id":              string   @go(TargetId) @yaml("target-id")
	"modification-type":      #ModType @go(ModType) @yaml("modification-type")
	"modification-rationale": string   @go(ModificationRationale) @yaml("modification-rationale")

	overrides?:  layer2.#Control    @go(Overrides,optional=nillable)
	extensions?: #ControlExtensions @go(Extensions,optional=nillable)
	"assessment-requirement-modifications"?: [...#AssessmentRequirementModifier] @go(AssessmentRequirementModifications) @yaml("assessment-requirement-modifications,omitempty")
}

#ControlExtensions: {
	severity?:                   #Severity @go(Severity)
	"auto-remediation-allowed"?: bool      @go(AutoRemediationAllowed) @yaml("auto-remediation-allowed,omitempty")
	"deployment-gate-allowed"?:  bool      @go(DeploymentGateAllowed) @yaml("deployment-gate-allowed,omitempty")
}

// Severity represents the severity level of a control
#Severity: "Critical" | "High" | "Medium" | "Low" | "Info" | "Unknown" @go(-)

#AssessmentRequirementModifier: {
	id?:                      string   @go(Id)
	"target-id":              string   @go(TargetId) @yaml("target-id")
	"modification-type":      #ModType @go(ModType) @yaml("modification-type")
	"modification-rationale": string   @go(ModificationRationale) @yaml("modification-rationale")

	overrides?:  layer2.#AssessmentRequirement    @go(Overrides,optional=nillable)
	extensions?: #AssessmentRequirementExtensions @go(Extensions,optional=nillable)
}

#AssessmentRequirementExtensions: {
	"required-evaluators"?: [...string] @go(RequiredEvaluators) @yaml("required-evaluators,omitempty")
	"optional-evaluators"?: [...string] @go(OptionalEvaluators) @yaml("optional-evaluators,omitempty")
	"evidence-requirements"?: string              @go(EvidenceRequirements) @yaml("evidence-requirements,omitempty")
	"resolution-strategy"?:   #ResolutionStrategy @go(ResolutionStrategy) @yaml("resolution-strategy,omitempty")
	"evaluation-points"?: [...#EvaluationPoint] @go(EvaluationPoints) @yaml("evaluation-points,omitempty")
}

#GuidelineModifier: {
	id?:                      string            @go(Id)
	"target-id":              string            @go(TargetId) @yaml("target-id")
	"modification-type":      #ModType          @go(ModType) @yaml("modification-type")
	"modification-rationale": string            @go(ModificationRationale) @yaml("modification-rationale")
	overrides?:               layer1.#Guideline @go(Overrides,optional=nillable)
	"statement-modifications"?: [...#StatementModifier] @go(StatementModifications) @yaml("statement-modifications,omitempty")
}

#StatementModifier: {
	id?:                      string            @go(Id)
	"target-id":              string            @go(TargetId) @yaml("target-id")
	"modification-type":      #ModType          @go(ModType) @yaml("modification-type")
	"modification-rationale": string            @go(ModificationRationale) @yaml("modification-rationale")
	overrides?:               layer1.#Statement @go(Overrides,optional=nillable)
}

#EvaluationPoint: "development-tools" |
	// For noncompliance risk to workflows or local machines
	"pre-commit-hook" |
	// For noncompliance risk to a development repository
	"pre-merge" |
	// For noncompliance risk to primary repositories
	"pre-build" |
	// For noncompliance risk to built assets
	"pre-release" |
	// For noncompliance risk to released assets
	"pre-deploy" |
	// For noncompliance risk to deployments
	"runtime-adhoc" |
	// For situations where drift may occur
	"runtime-scheduled" |
	// For situations where drift detection is automated
	"runtime-reactive"
// For situations where drift detection is triggered by events

#EnforcementMethod: "Deployment Gate" |
	"Autoremediation" |
	"Manual Remediation"

#NotificationGroup: "Responsible" |
	"Accountable" |
	"Consulted" |
	"Informed"

#ModType: "include" | "increase-strictness" | "clarify" | "reduce-strictness" | "exclude"

// ResolutionStrategy defines how to resolve conflicts when multiple evaluators produce different results
// for the same assessment requirement. Options:
// - "MostSevere": Use the most severe result from all evaluators (Failed > Unknown > NeedsReview > Passed)
// - "ManualOverride": Give precedence to manual review evaluators over automated evaluators when results conflict
// - "AuthoritativeConfirmation": Require confirmation from authoritative evaluators before triggering findings from non-authoritative evaluators
#ResolutionStrategy: "MostSevere" | "ManualOverride" | "AuthoritativeConfirmation" @go(-)

// ImportedPolicy represents a reference to another policy that is imported.
#ImportedPolicy: {
	"reference-id": string @go(ReferenceId) @yaml("reference-id")
}
