// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// AuditLog records results from an audit performed against a target resource
#AuditLog: {
	#Log
	metadata: type: "AuditLog"

	// owner defines the RACI roles responsible for managing the audit
	owner?: #RACI @go(Owner)

	// summary provides the high-level conclusion
	summary: string

	// policy is the single policy this audit was conducted against, at its
	// pinned version. Auditing a baseline means authoring a policy that
	// imports it.
	policy: #ArtifactMapping

	// coverage records, for every requirement the evaluation resolved, what
	// this audit checked and how it came out. It is complete by construction:
	// one entry per effective requirement, including the boring passes.
	coverage: [#RequirementCoverage, ...#RequirementCoverage]

	// results are the auditor's opinions, which are selective by nature.
	results: [#AuditResult, ...#AuditResult] @go(Results,type=[]*AuditResult)

	// risk-tolerance-met records the policy-level tolerance check, which is not
	// per-requirement.
	"risk-tolerance-met"?: #CheckOutcome @go(RiskToleranceMet)

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// Every reference-id names a declared mapping-reference or the audit's own
	// id: the audited policy, each covered requirement, each evidence source.

	_refIds: _ // computed on #Log; named here only so this block can reference it

	_refValidation: "policy": _refIds & list.Contains(policy."reference-id")
	for i, c in coverage {
		_refValidation: "coverage-\(i)": _refIds & list.Contains(c.requirement."reference-id")
		if c.evidence != _|_ {
			for j, e in c.evidence if e.source != _|_ {
				_refValidation: "coverage-\(i)-evidence-\(j)": _refIds & list.Contains(e.source."reference-id")
			}
		}
	}

	// Coverage ids are unique, because a result addresses one by check-id.

	_uniqueCoverageIds: {for i, c in coverage {(c.id): i}}

	// A result's check-id, when present, names one of this audit's coverage
	// entries, so an opinion and the check that prompted it cannot disagree.

	_coverageIds: [for c in coverage {c.id}]
	for i, r in results if r."check-id" != _|_ {
		_checkValidation: "\(i)": _coverageIds & list.Contains(r."check-id")
	}
}

// ResultType classifies the nature of an audit result
#ResultType: "Gap" | "Finding" | "Observation" | "Strength" @go(-)

// CheckOutcome is the result of one audit check.
#CheckOutcome: {
	outcome:  "Pass" | "Fail" | "Not Applicable" | "Undetermined"
	message?: string
}

// AuditResult records a single result with supporting evidence and recommendations.
#AuditResult: {
	// id uniquely identifies this result
	id: string

	// title describes this result at a glance
	title: string

	// type classifies the nature of this result
	type: #ResultType

	// description explains the result in detail
	description: string

	// check-id names the coverage entry whose check prompted this result, when
	// one did. A result about the audit as a whole omits it rather than
	// inventing a check to point at.
	"check-id"?: string @go(CheckId)

	// recommendations records corrective actions for this result
	recommendations?: [#Recommendation, ...#Recommendation] @go(Recommendations)
}

// Recommendation provides a corrective action for an audit result
#Recommendation: {
	// id uniquely identifies this recommendation
	id?: string

	// text describes the recommended corrective action
	text: string

	// required indicates whether this recommendation is a mandatory corrective action
	required: *false | bool @gemara(default=false)
}

// RequirementCoverage records what an audit checked for one requirement. Each
// check is a named field so that a check never made is visible as a missing
// field rather than an absent list entry.
#RequirementCoverage: {
	id: string
	// requirement names an entry of the evaluation log's effective-requirements.
	requirement: #EntryMapping

	// assessments are the assessments this audit read, if any.
	assessments?: [#EvidenceMapping, ...#EvidenceMapping]

	// evidence records the data sources that support this result
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)

	// assessment-present records whether the requirement was assessed at all.
	// It is required: it is what makes coverage complete rather than selective.
	"assessment-present": #CheckOutcome @go(AssessmentPresent)

	// plan-bound records whether the assessment ran under the plan the policy
	// binds to this requirement.
	"plan-bound"?: #CheckOutcome @go(PlanBound)

	// method-allowed records whether the method that ran is one the plan allows.
	"method-allowed"?: #CheckOutcome @go(MethodAllowed)

	// executor-matched records whether the executor matched the plan's, compared
	// by authoritative identifier rather than by internal short name.
	"executor-matched"?: #CheckOutcome @go(ExecutorMatched)

	// parameters-matched records whether the recorded parameter values are the
	// ones the plan declares, within their accepted values.
	"parameters-matched"?: #CheckOutcome @go(ParametersMatched)

	// frequency-met records whether each required method ran within the
	// requirement's frequency-days.
	"frequency-met"?: #CheckOutcome @go(FrequencyMet)

	// evidence-present records whether the requirement's assessments carry at
	// least one evidence entry.
	"evidence-present"?: #CheckOutcome @go(EvidencePresent)

	// evidence-fresh records whether that evidence was collected within the
	// plan's valid-for-days.
	EF="evidence-fresh"?: #CheckOutcome @go(EvidenceFresh)

	// required-methods records one outcome per method the plan requires.
	"required-methods"?: [...{"method-id": string, outcome: #CheckOutcome}] @go(RequiredMethods)

	// ---- Validation --------------------------------------------------------

	// Each evidence entry carries a payload, a source, or both.

	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]

	// evidence-fresh only means something once evidence was found. It is read
	// through the EF alias because a quoted name in a CUE expression is a
	// string literal, not a field, so testing it directly always succeeds.

	if EF != _|_ {"evidence-present": #CheckOutcome}
}
