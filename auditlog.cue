// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// Determination is what a named check in a verification found: whether the
// condition the check names holds. It is deliberately not #Result's vocabulary —
// #Result says whether an assessment procedure passed when it ran — and
// deliberately not #ComplianceStatus: a check asserts a condition, while
// compliance is what an audit concludes from the checks and the evidence
// together. Undetermined is the word #Disposition already uses for "could not be
// told".
#Determination: "Satisfied" | "Not Satisfied" | "Not Applicable" | "Undetermined" @go(-)

// ComplianceStatus is whether the target complies with one requirement. It is the
// audit's conclusion, which an auditor may reach differently from what the
// assessment observed, and it is what projects to OCSF's compliance.status.
// Compliance is stated one requirement at a time: there is no "compliant with
// exceptions" here, because weighing exceptions is a judgement about the target as
// a whole, which belongs to #Opinion.
#ComplianceStatus: "Compliant" | "Not Compliant" | "Not Applicable" | "Undetermined" @go(-)

// Opinion is the audit's conclusion about the target as a whole: ISO 19011's audit
// conclusion, reached after weighing all of the findings against what the audit set
// out to establish. It carries the
// audit profession's four opinions: unqualified (Passed), qualified (Passed with
// Conditions), adverse (Failed), and a disclaimer of opinion (Undetermined), when
// the auditor could not obtain enough to conclude. It borrows #Result's tense
// because it answers the same shape of question at the scale of a whole audit.
//
// It is a judgement, not a roll-up of verifications: whether requirements that
// were not met are material enough to sink the audit is the auditor's call, and no
// precedence over per-requirement outcomes can compute it. That is why it is
// stated rather than derived.
#Opinion: "Passed" | "Passed with Conditions" | "Failed" | "Undetermined" @go(-)

// AuditLog records results from an audit performed against a target resource
#AuditLog: {
	#Log
	metadata: type: "AuditLog"

	// owner defines the RACI roles responsible for managing the audit
	owner?: #RACI @go(Owner,optional=nillable)

	// summary provides the high-level conclusion
	summary: string

	// opinion is that same conclusion in one word, so a consumer of the
	// attestation need not weigh the findings itself. summary carries the prose;
	// this is the machine-readable one. Anything other than Passed is expected to
	// be explained by findings, documented rather than enforced.
	opinion: #Opinion

	// policy is the single policy this audit was conducted against, at its
	// pinned version. Auditing a baseline means authoring a policy that
	// imports it.
	policy: #ArtifactMapping

	// verifications records, for every requirement, what was mechanically checked
	// and how it came out. It is complete by construction — one entry per
	// requirement, including the boring passes — and carries no reasoning: it is
	// the audit's counterpart of #ControlEvaluation, not of #ComplianceFinding.
	verifications: [#VerificationLog, ...#VerificationLog]

	// findings are what this audit reports: the called-out determinations, each
	// naming the requirement or the risk it concerns. verifications is the
	// systematic view, one entry per requirement; this is the selective one, and
	// the only one that reasons. It is absent when the audit had nothing to call
	// out, which is what a Passed opinion looks like.
	findings?: [...#ComplianceFinding] @go(Findings,type=[]*ComplianceFinding)

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// Every reference-id names a declared mapping-reference or the audit's own
	// id: the audited policy, each covered requirement, each evidence source.

	// A finding the audit determines Not Compliant states how material it is. An
	// unrated Not Compliant finding cannot be triaged by anything downstream, which
	// is most of what a compliance finding is for. The rule is written as an
	// element constraint rather than as a conditional inside #ComplianceFinding,
	// because a conditional on a required enum leaves the definition itself
	// unevaluable.

	findings?: [...{
		status: #ComplianceStatus // declared so the reference below resolves to this element

		if status == "Not Compliant" {
			severity: #Severity
		}
	}]

	_refIds: _ // computed on #Log; named here only so this block can reference it

	_refValidation: "policy": _refIds & list.Contains(policy."reference-id")
	for i, c in verifications {
		_refValidation: "verifications-\(i)": _refIds & list.Contains(c.requirement."reference-id")
		if c.evidence != _|_ {
			for j, e in c.evidence if e.source != _|_ {
				_refValidation: "verifications-\(i)-evidence-\(j)": _refIds & list.Contains(e.source."reference-id")
			}
		}
	}

	if findings != _|_ {
		for i, f in findings if f.risk != _|_ {
			_refValidation: "findings-\(i)-risk": _refIds & list.Contains(f.risk."reference-id")
		}

		for i, f in findings if f.requirement != _|_ {
			_refValidation: "findings-\(i)-requirement": _refIds & list.Contains(f.requirement."reference-id")
		}
	}

	// A finding names either a requirement this audit verified, or a risk. Naming a
	// requirement with no verification entry would report on something the audit
	// never checked.

	_verifiedRequirements: [for c in verifications {"\(c.requirement."reference-id")/\(c.requirement."entry-id")"}]
	if findings != _|_ {
		for i, f in findings if f.requirement != _|_ {
			_requirementValidation: "\(i)": _verifiedRequirements & list.Contains("\(f.requirement."reference-id")/\(f.requirement."entry-id")")
		}
	}
}

// PlanVerification records whether the plan bound to a requirement was followed.
// Method-specific checks sit per method, because a plan may require several
// evaluation methods and each produces its own assessment: a single outcome per
// requirement could not say which method fell short.
#PlanVerification: {
	// bound records whether the assessment ran under the plan the policy binds to
	// this requirement.
	bound: #Determination

	// methods records one entry per method involved: those the plan requires, and
	// any that ran without the plan allowing them. At least one, because a plan
	// declares at least one accepted method, so a plan verification that
	// enumerates none has verified nothing about the methods.
	methods: [#MethodVerification, ...#MethodVerification]
}

// MethodVerification records how one evaluation method fared against its plan.
#MethodVerification: {
	// method-id names the accepted method this entry reports on.
	"method-id": string @go(MethodId)

	// allowed records whether the plan permits this method. It is required: a
	// method entry can always answer it, and an entry that names a method without
	// saying whether the plan accepts it has verified nothing.
	allowed: #Determination

	// used records whether this method was actually used. It is required for the
	// same reason as allowed: whether a method ran is always answerable, and a
	// required method that did not run is the failure this list exists to surface.
	used: #Determination

	// executor-matched records whether the executor matched the plan's, compared
	// by authoritative identifier rather than by internal short name.
	"executor-matched"?: #Determination @go(ExecutorMatched)

	// parameters-matched records whether the recorded parameter values are the
	// ones the plan declares, within their accepted values.
	"parameters-matched"?: #Determination @go(ParametersMatched)

	// cadence-met records whether this method ran within the requirement's
	// frequency-days. It is required because every plan states a frequency, so
	// there is always a cadence to have met; a method that ran without the plan
	// allowing it has none to meet, which is Not Applicable. It is recorded per
	// method rather than per requirement so that a lagging method is named; the
	// requirement's answer is the outcome above.
	"cadence-met": #Determination @go(CadenceMet)
}

// EffectiveRequirement is what a policy required of one requirement, as resolved
// at audit time: the text as it applied, the applicability that selected it, the
// constraints the policy attached, and the plan bound to it. It is copied into
// the audit on purpose. An audit is a point-in-time attestation, and a pointer
// alone stops saying what was audited once the catalog moves, retires the
// requirement, or the policy's constraints change.
#EffectiveRequirement: {
	// text is the requirement as it applied, after the policy's constraints.
	text: string

	// applicability names the groups under which this requirement was selected.
	applicability?: [string, ...string]

	// constraints are the prescriptive additions the policy attached to it.
	constraints?: [...{id: string, text: string}]

	// plan-id names the assessment plan bound to this requirement.
	"plan-id"?: string @go(PlanId)
}

// ComplianceFinding is what an audit reports in its own right: a finding carrying
// the audit's compliance determination. It is the only record in Gemara that
// decides compliance, and the audit's counterpart of #ActionResult — the record
// type this log carries a list of.
#ComplianceFinding: {
	#Finding

	// status is the audit's compliance determination for what this finding
	// concerns. Only the audit decides whether a requirement is satisfied: an
	// evaluation says whether its procedure passed, an enforcement action says
	// what it did. It carries OCSF's compliance.status.
	status: #ComplianceStatus

	// id uniquely identifies this finding within the audit. It is declared here
	// rather than in the #Finding core because a finding reported against a named
	// check needs no id of its own, while this one is reported in its own right.
	id: string

	// risk names the risk this finding concerns, when it concerns one: a residual
	// risk beyond its category's tolerance, for example. A finding names a risk or
	// a requirement — the requirement comes from the #Finding core.
	risk?: #EntryMapping @go(Risk,optional=nillable)
}

// VerificationLog records what an audit mechanically checked for one requirement,
// and how it came out. It is the audit's counterpart of #AssessmentLog: a record
// of a check that ran, carrying an outcome and no reasoning. Each check is a named
// field so that a check never made is visible as a missing field rather than as an
// absent list entry.
#VerificationLog: {
	// outcome is whether the target complies with this requirement. It is stated
	// rather than left to be recomputed from the checks below or from the
	// findings: giving every requirement an answer is what this layer is for.
	// A check that came out Not Satisfied is expected to show here, since a
	// requirement that is unevidenced, stale or unverified against its plan
	// cannot be determined Satisfied. An outcome a reader would question is
	// explained by a finding naming this requirement, because nothing recorded
	// here reasons. Both relationships are documented rather than enforced, as
	// #Result's roll-up precedence is.
	outcome: #ComplianceStatus

	// requirement names the assessment requirement this entry determines, and is
	// also this entry's identity: one entry per requirement, so a finding names
	// the requirement rather than a minted id.
	requirement: #EntryMapping

	// effective is what the policy actually required of it, resolved at audit
	// time, so the record says what was audited without fetching the catalog.
	effective: #EffectiveRequirement

	// assessments are the assessments this audit read, if any.
	assessments?: [#EvidenceMapping, ...#EvidenceMapping]

	// evidence records the data sources that support this result
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)

	// evidence-present records whether this requirement is backed by evidence at
	// all: the question verifications exist to answer, and answering it for every
	// requirement is what makes them complete rather than selective. It may be
	// omitted only where there was no evidence question to ask, which is a
	// requirement the policy puts out of scope — and omitting it then obliges an
	// outcome of Not Applicable, so the omission cannot hide an entry that was
	// simply left unanswered.
	EP="evidence-present"?: #Determination @go(EvidencePresent)

	// evidence-fresh records whether that evidence is current: collected within
	// the plan's valid-for-days where a plan sets one, and within the auditor's
	// judgement where none does.
	"evidence-fresh"?: #Determination @go(EvidenceFresh)

	// plan records whether the plan behind this requirement was followed. It is
	// absent when no plan backed the requirement, which is how a verification says
	// the requirement was evidenced without one, and required when effective
	// names a plan-id.
	plan?: #PlanVerification @go(Plan,optional=nillable)

	// ---- Validation --------------------------------------------------------

	// A named plan obliges a record of how it was followed, and a record of how a
	// plan was followed obliges the plan-id it was verified against. This is the
	// obligation #AssessmentLog puts on plan and execution, at the checking end of
	// the same chain.
	//
	// Both directions are stated positively, on the field each requires. Forbidding
	// a field instead — plan?: error(...) where no plan-id is named — puts bottom in
	// that field's type, and the Go projection degrades it to `any` with a TODO,
	// which takes the field away from every consumer to catch an authoring mistake.

	if effective."plan-id" != _|_ {plan: #PlanVerification}

	if plan != _|_ {effective: "plan-id": string}

	// Omitting evidence-present says there was no evidence question to ask, which is
	// true only of a requirement the audit determined Not Applicable. Everything the
	// audit actually examined answers it.

	if EP == _|_ {
		outcome: "Not Applicable"
	}

	// An entry that names nothing cannot claim coverage. Stated as the
	// contrapositive — no evidence and no assessments means evidence-present is not
	// Satisfied — because conditioning on the absence of optional fields is
	// decidable, while conditioning on the value of a required enum leaves the
	// definition unevaluable and reports the wrong field when one is missing.
	// Note that == _|_ is true of an invalid value as well as an absent one, so an
	// entry whose evidence fails its own rules also reports here; the entry has no
	// valid evidence either way. The constraint is optional so that it narrows the
	// value when one is given without forcing the field back into existence for an
	// out-of-scope requirement that omits it.

	if evidence == _|_ if assessments == _|_ {
		"evidence-present"?: "Not Satisfied" | "Not Applicable" | "Undetermined"
	}

	// Each evidence entry carries a payload, a source, or both.

	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]
}

// Evidence records what was cited to support an opinion for a specific activity:
// raw data for the evaluation layer, evaluation and enforcement artifacts for the audit layer.
// At least one of payload or source MUST be present; an entry with neither is semantically incomplete.
#Evidence: {
	// id uniquely identifies this evidence
	id: string

	// type categorizes the kind of evidence
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// payload is the raw evidence data collected inline
	payload?: _ @go(Payload,type=any)

	// source identifies the artifact or system from which this evidence was collected
	source?: #EvidenceMapping @go(Source)

	// description explains what this evidence represents
	description?: string
}

// _EvidenceStrict layers the "at least one of payload or source" rule on top of #Evidence
#_EvidenceStrict: {
	@go(-)
} & #Evidence & {
	payload?: _
	if payload == _|_ {
		source: #EvidenceMapping
	}
}

// EvidenceType categorizes the kind of evidence. It remains an open enum:
// recommended values include artifact types already known to Gemara (e.g.
// EvaluationLog, EnforcementLog) plus categories for common evidence forms.
#EvidenceType: #ArtifactType | string @go(-)
