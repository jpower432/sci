// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// Policy represents a policy document with metadata, contacts, scope, imports, implementation plan, risks, and adherence requirements.
#Policy: {
	title:    string
	metadata: #Metadata
	metadata: type: "Policy"
	contacts:               #RACI
	scope:                  #Scope
	imports:                #Imports
	"implementation-plan"?: #ImplementationPlan @go(ImplementationPlan)
	risks?:                 #Risks
	adherence:              #Adherence

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// _refIds is every id a reference-id in this policy may name: its declared
	// mapping-references, plus its own metadata.id. A Policy is not a #Log or
	// #Catalog, so it computes its own. It is a hidden field rather than a `let`
	// because the OpenAPI converter strips the comprehensions that read it, and a
	// `let` left unreferenced by that pass is an error.

	_refIds: list.Concat([[metadata.id], [
		if metadata["mapping-references"] != _|_
		for r in metadata["mapping-references"] {r.id},
	]])

	// Every reference-id in this policy's imports names a declared reference.

	if imports.policies != _|_ {
		for i, p in imports.policies {
			_refValidation: "imports-policies-\(i)": _refIds & list.Contains(p."reference-id")
		}
	}
	if imports.catalogs != _|_ {
		for i, c in imports.catalogs {
			_refValidation: "imports-catalogs-\(i)": _refIds & list.Contains(c."reference-id")
		}
	}

	if imports.guidance != _|_ {
		for i, g in imports.guidance {
			_refValidation: "imports-guidance-\(i)": _refIds & list.Contains(g."reference-id")
		}
	}

	// Plan ids are unique: AssessmentLog.plan and AssessmentFinding.plan
	// address a plan by this id.

	// A policy binds one plan per requirement: two plans governing the same
	// requirement would leave an evaluation with no way to choose, and the audit's
	// effective.plan-id with nothing single to name.

	if adherence."assessment-plans" != _|_ {
		_uniquePlanRequirements: {
			for i, p in adherence."assessment-plans" {
				"\(p.requirement."reference-id")/\(p.requirement."entry-id")": i
			}
		}
	}

	// Each plan's requirement names a catalog this policy declares.

	if adherence."assessment-plans" != _|_ {
		for i, p in adherence."assessment-plans" {
			_refValidation: "assessment-plans-\(i)-requirement": _refIds & list.Contains(p.requirement."reference-id")
		}
	}

	if adherence."assessment-plans" != _|_ {
		_uniquePlanIds: {for i, p in adherence."assessment-plans" {(p.id): i}}
	}

	// Accepted methods share one id space across all three sites they can appear
	// in, because ActionResult.method is a bare EntryMapping with no way to say
	// which site it means. The value records the site, so a collision names both
	// positions; a per-list check would miss a cross-site duplicate entirely.

	_uniqueMethodIds: {
		if adherence."evaluation-methods" != _|_ {
			for i, m in adherence."evaluation-methods" {(m.id): "evaluation-methods[\(i)]"}
		}
		if adherence."assessment-plans" != _|_ {
			for i, p in adherence."assessment-plans" for j, m in p."evaluation-methods" {
				(m.id): "assessment-plans[\(i)].evaluation-methods[\(j)]"
			}
		}
		if adherence."enforcement-methods" != _|_ {
			for i, m in adherence."enforcement-methods" {(m.id): "enforcement-methods[\(i)]"}
		}
	}
}

// Scope defines what is included and excluded from policy applicability.
#Scope: {
	in:   #Dimensions
	out?: #Dimensions
}

// Dimensions specify the applicability criteria for a policy
#Dimensions: {
	// technologies is an optional list of technology categories or services
	technologies?: [string, ...string]
	// geopolitical is an optional list of geopolitical regions
	geopolitical?: [string, ...string]
	// sensitivity is an optional list of data classification levels
	sensitivity?: [string, ...string]
	// users is an optional list of user roles
	users?: [string, ...string]
	groups?: [string, ...string]
}

// Imports defines external policies, controls, and guidelines required by this policy.
#Imports: {
	policies?: [#ArtifactMapping, ...#ArtifactMapping]
	catalogs?: [#CatalogImport, ...#CatalogImport]
	guidance?: [#GuidanceImport, ...#GuidanceImport]
}

// ImplementationPlan defines when and how the policy becomes active. Its timelines
// are communicated rather than enforced: no log references them, and no rule here
// compares a log's timestamps against them. They carry interpretation a later reader
// cannot reconstruct — that enforcement was not yet due rather than skipped, and
// whether an audit's conformance statement falls inside the window the policy was
// actually in effect. Comparing a log or an audit period against them belongs to a
// tool, which has both documents in hand.
#ImplementationPlan: {
	"notification-process"?: string                 @go(NotificationProcess)
	"evaluation-timeline":   #ImplementationDetails @go(EvaluationTimeline)
	"enforcement-timeline":  #ImplementationDetails @go(EnforcementTimeline)
}

// ImplementationDetails specifies one timeline for policy implementation: when it
// starts, when it ends if it does, and what a reader should know about it.
#ImplementationDetails: {
	start: #Datetime
	end?:  #Datetime
	notes: string
}

// Risks defines mitigated and accepted risks addressed by this policy.
#Risks: {
	// Mitigated risks only need reference-id and risk-id (no justification required)
	mitigated?: [#MitigatedRisk, ...#MitigatedRisk]
	// Accepted risks require rationale (justification) and may include scope. Controls addressing these risks are implicitly identified through threat mappings.
	accepted?: [#AcceptedRisk, ...#AcceptedRisk]
}

// MitigatedRisk represents a risk addressed by the policy
#MitigatedRisk: {
	// id allows this mitigated risk entry to be referenced by accepted risks
	id: string

	// risk references the risk being mitigated
	risk: #EntryMapping
}

// AcceptedRisk documents a risk the organization has chosen to accept,
// optionally linking it to a mitigated risk when the acceptance covers
// residual risk after partial mitigation.
#AcceptedRisk: {
	// id allows this accepted risk entry to be referenced
	id: string

	// target-id optionally links this acceptance to a mitigated risk entry
	"target-id"?: string

	// risk references the risk being accepted
	risk: #EntryMapping

	// scope defines where the risk acceptance applies
	scope?: #Scope

	// justification explains why the risk is accepted
	justification?: string
}

// Adherence defines evaluation methods, assessment plans, enforcement methods, and non-compliance notifications.
#Adherence: {
	"evaluation-methods"?: [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)
	"assessment-plans"?: [#AssessmentPlan, ...#AssessmentPlan] @go(AssessmentPlans)
	"enforcement-methods"?: [#AcceptedMethod & {type: #EnforcementMethodType}, ...#AcceptedMethod & {type: #EnforcementMethodType}] @go(EnforcementMethods)
	"non-compliance"?: string @go(NonCompliance)
}

// AssessmentPlan defines how a specific assessment requirement is evaluated.
#AssessmentPlan: {
	id: string

	// requirement names the assessment requirement this plan governs, in the
	// catalog it comes from.
	requirement: #EntryMapping @go(Requirement)

	// frequency-days is how often this requirement is re-examined, in days.
	// Cadence belongs to the requirement, not to a method or a plan instance.
	"frequency-days": int & >0 @go(FrequencyDays)

	"evaluation-methods": [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)

	// evidence-requirements states what evidence this requirement needs and how
	// long a collected item stays usable. It declares no evidence types: an
	// expectation belongs to a policy and a requirement.
	"evidence-requirements"?: #EvidenceRequirements @go(EvidenceRequirements)

	parameters?: [#Parameter, ...#Parameter]

	// ---- Validation --------------------------------------------------------

	// Parameter ids are unique within their plan, so an executor resolving a
	// parameter by id gets one value.

	if parameters != _|_ {
		_uniqueParameterIds: {for i, p in parameters {(p.id): i}}
	}
}

// EvidenceRequirements states what evidence a requirement needs and how long a
// collected item stays usable. valid-for-days is a different clock from
// frequency-days: how long evidence stays probative, not how often the
// requirement is re-examined.
#EvidenceRequirements: {
	description:      string
	"valid-for-days": int & >0 @go(ValidForDays)
}

// AcceptedMethod defines a method for evaluation or enforcement.
#AcceptedMethod: {
	id:           string
	type:         #MethodType
	mode:         #ModeType
	required:     *false | bool @gemara(default=false)
	description?: string
	executor?:    #Actor
}

#ModeType:              "Manual" | "Automated"                           @go(-)
#MethodType:            "Behavioral" | "Intent" | "Remediation" | "Gate" @go(-)
#EvaluationMethodType:  "Intent" | "Behavioral"                          @go(-)
#EnforcementMethodType: "Gate" | "Remediation"                           @go(-)

// Parameter defines a configurable parameter for assessment or enforcement activities.
#Parameter: {
	id:          string
	label:       string
	description: string
	"accepted-values"?: [string, ...string] @go(AcceptedValues)
}

// GuidanceImport defines how to import guidance documents with optional exclusions and constraints.
#GuidanceImport: {
	"reference-id": string @go(ReferenceId)
	exclusions?: [string, ...string]
	// Constraints allow policy authors to define ad hoc minimum requirements (e.g., "review at least annually").
	constraints?: [#Constraint, ...#Constraint]
}

// CatalogImport defines how to import control catalogs with optional exclusions, constraints, and assessment requirement modifications.
#CatalogImport: {
	"reference-id": string @go(ReferenceId)
	exclusions?: [string, ...string]
	constraints?: [#Constraint, ...#Constraint]
	"assessment-requirement-modifications"?: [#AssessmentRequirementModifier, ...#AssessmentRequirementModifier] @go(AssessmentRequirementModifications)
}

// Constraint defines a prescriptive requirement that applies to a specific guidance or control.
#Constraint: {
	// Unique ID for this constraint to enable Layer 5/6 tracking
	id: string
	// Links to the specific Guidance or Control being constrained
	"target-id": string @go(TargetId)
	// The prescriptive requirement/constraint text
	text: string
}

// AssessmentRequirementModifier allows organizations to customize assessment requirements based on how an organization wants to gather evidence for the objective.
#AssessmentRequirementModifier: {
	id:                       string
	"target-id":              string   @go(TargetId)
	"modification-type":      #ModType @go(ModificationType)
	"modification-rationale": string   @go(ModificationRationale)
	// The updated text of the assessment requirement
	text?: string
	// The updated applicability of the assessment requirement
	applicability?: [string, ...string]
	// The updated recommendation for the assessment requirement
	recommendation?: string
}

// ModType defines the type of modification to the assessment requirement.
#ModType: "Add" | "Modify" | "Remove" | "Replace" | "Override" @go(-)
