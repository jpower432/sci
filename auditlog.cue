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

// Opinion is the audit's conclusion about the target as a whole: an audit
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

// AuditLog records an audit: what it examined the target against, what it
// established for each requirement, what it reports, and what it concludes. The
// fields follow the order the work happens in — criteria, then verification, then
// synthesis, then decision — which is also ISO 19011's sequence: collecting and
// verifying information, generating findings, determining conclusions.
#AuditLog: {
	#Log
	metadata: type: "AuditLog"

	// owner defines the RACI roles responsible for managing the audit
	owner?: #RACI @go(Owner,optional=nillable)

	// policy is this audit's criteria: the single versioned artifact the evidence
	// was compared against. It is one policy rather than a
	// list because the policy is also what defines scope — selection, exclusions
	// and applicability — and the audit's completeness claim has nothing to be
	// complete against without a governed set. Auditing a baseline means authoring
	// a policy that imports it; auditing against several criteria means one policy
	// that imports them all, and importing guidance into a policy makes it binding.
	//
	// This does not hide the criteria's content: the underlying catalogs stay named
	// in metadata.mapping-references, and each verification carries the requirement
	// as resolved, so a reader sees what was required without fetching the policy.
	// It does mean the bare plan-id references elsewhere in this file are legal only
	// while the criteria is singular — with one policy pinned here, nothing else
	// needs a qualifier.
	policy: #ArtifactMapping

	// objectives are what this audit set out to establish. They are stated because
	// the conclusion has to speak to them: an opinion is an answer to a question,
	// and without the question recorded a later reader cannot tell whether the audit
	// achieved what it was for, or judge the scope it settled for.
	//
	// These are the audit's own objectives, not the mandates the policy adopted —
	// whether the policy implements those is recorded in coverage.
	objectives: [string, ...string]

	// verifications records, for every requirement the policy governs, what was
	// examined and what it established. It is complete by construction — one entry
	// per requirement, including the boring passes — and carries determinations
	// rather than judgements.
	verifications: [#VerificationLog, ...#VerificationLog]

	// coverage records, for every mandate the policy adopted from guidance or a
	// parent policy, whether the policy implements it. It is complete over those
	// mandates for the same reason verifications is complete over requirements: an
	// audit that reported the gaps it happened to notice could not be shown to have
	// looked for the rest. Control objectives are absent by design — their
	// assessment requirements determine them, so selecting the requirements is
	// implementing the objective.
	//
	// Whether a mandate is met is not here. Its statements are governed requirements
	// with determinations of their own, and a mandate with no statements is itself a
	// governed requirement with a verification entry.
	//
	// Not to be confused with objectives above, which are the audit's own — what it
	// set out to establish, rather than what the policy adopted from its imports.
	coverage: [#MandateCoverage, ...#MandateCoverage]

	// findings are what this audit reports: the synthesis over those verifications,
	// each naming the requirement or the risk it concerns and carrying the
	// compliance judgement that the verification layer deliberately does not make.
	// It is absent when the audit had nothing to call out, which is what a Passed
	// opinion looks like.
	findings?: [...#ComplianceFinding] @go(Findings,type=[]*ComplianceFinding)

	// opinion is the audit's conclusion in one word, so a consumer of the
	// attestation need not weigh the findings itself.
	opinion: #Opinion

	// summary is that conclusion in prose, including anything the opinion cannot
	// carry: what the objectives established, trends across the findings, and the
	// uncertainty any audit of a moment carries.
	summary: string

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

	// A coverage entry names a declared reference, and what it says implements the
	// mandate is something this audit actually verified.

	for i, c in coverage {
		_refValidation: "coverage-\(i)": _refIds & list.Contains(c.mandate."reference-id")

		if c["implemented-by"] != _|_ {
			for j, r in c["implemented-by"] {
				_requirementValidation: "coverage-\(i)-\(j)": _verifiedRequirements & list.Contains("\(r."reference-id")/\(r."entry-id")")
			}
		}
	}

	// A requirement the evidence did not satisfy must be synthesised. Without a
	// finding naming it, the record says the requirement was not met and never says
	// what that means — non-compliance, or a gap the auditor accepted — which is
	// exactly the judgement the verification layer defers to this one.

	_findingRequirements: [
		if findings != _|_
		for f in findings if f.requirement != _|_ {"\(f.requirement."reference-id")/\(f.requirement."entry-id")"},
	]

	// Written as an assignment rather than a conditional: an incomplete value is
	// tolerated on the right of an assignment and fatal in an if clause, and
	// outcome is a disjunction until the data makes it concrete. A conditional
	// here leaves #AuditLog itself unevaluable.

	for i, v in verifications {
		_synthesisValidation: "\(i)": true & (v.outcome != "Not Satisfied" ||
			list.Contains(_findingRequirements, "\(v.requirement."reference-id")/\(v.requirement."entry-id")"))
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
	constraints?: [...#EffectiveConstraint]

	// plan-id names the assessment plan bound to this requirement.
	"plan-id"?: string @go(PlanId)
}

// EffectiveConstraint is one prescriptive addition a policy attached to a
// requirement, as it applied. It is a named type rather than an inline struct so
// that consumers can name it: an anonymous struct projects into Go as an anonymous
// struct, which nothing can take as an argument or build a literal of without
// repeating the definition. It carries no target-id, unlike the policy's
// #Constraint, because the constraint is already attached to the requirement it
// targets.
#EffectiveConstraint: {
	// id identifies this constraint within the policy that attached it.
	id: string

	// text is the constraint as it applied to the requirement.
	text: string
}

// MandateCoverage records whether the policy implements a mandate it adopted from
// guidance or a parent policy. It is the one question about the criteria that
// nothing else answers: the verifications are each about a requirement the policy
// selected, so every one of them can be satisfied while a mandate the policy
// adopted goes unaddressed, and no roll-up over them can show it.
//
// Whether the mandate is *met* is not recorded here. A mandate's statements are
// governed requirements like any other, so their determinations say it; a mandate
// with no statements is itself the governed requirement and has a verification
// entry of its own. One mechanical determination per governed thing, and this is
// not one — it is about what the policy requires, not about what the target did.
#MandateCoverage: {
	// mandate names the guideline or parent-policy entry the policy adopted.
	mandate: #EntryMapping

	// statement is that mandate as the policy adopted it, resolved at audit time for
	// the same reason effective.text is: a pointer stops saying what was adopted
	// once the catalog moves.
	statement: string

	// implementation is whether the policy's selections implement it. Not Applicable
	// where the policy excluded it deliberately, which a tool holding the policy can
	// check against its exclusion list.
	implementation: #Determination

	// implemented-by names what in the policy does the work: requirements it
	// selected, or constraints it added. Absent alongside Not Satisfied, it is the
	// visible form of a mandate nothing implements; present alongside Not Satisfied,
	// it is partial coverage, with basis saying what is missing.
	"implemented-by"?: [#EntryMapping, ...#EntryMapping] @go(ImplementedBy)

	// basis is why. Nothing mechanical establishes this, so a bare determination
	// would be an assertion.
	basis?: string
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

// VerificationLog records what an audit established for one requirement: what it
// examined, whether the evidence was obtained the way the policy required, and
// whether the requirement was satisfied by it. It is the audit's counterpart of
// #AssessmentLog — a record of checking, carrying determinations and no judgement.
//
// Its vocabulary is deliberately #Determination rather than #ComplianceStatus.
// Conformance is about satisfaction, which is mechanical; compliance is a judgement,
// and judgement belongs to the findings that synthesise these entries. A reader who
// wants to know what was met reads here; a reader who wants to know what it means
// reads a finding.
#VerificationLog: {
	// requirement names the assessment requirement this entry determines, and is
	// also this entry's identity: one entry per requirement, so a finding names the
	// requirement rather than a minted id.
	requirement: #EntryMapping

	// effective is what the policy actually required of it, resolved at audit time,
	// so the record says what was audited without fetching the catalog. Importing
	// guidance into a policy makes it binding, so an effective requirement may have
	// come from a control catalog, from guidance, or from another policy.
	effective: #EffectiveRequirement

	// assessments are the assessments this audit read, if any.
	assessments?: [#EvidenceMapping, ...#EvidenceMapping]

	// evidence records the data sources that support this determination.
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)

	// plan-conformance records whether the evidence was obtained the way the policy
	// required: under the plan bound to this requirement, by one of its accepted
	// methods, by the executor it names, within its cadence. It is one determination
	// rather than a field per dimension, because the policy grows dimensions and a
	// record that mirrored them would have to grow with it — and because the detail
	// of a shortfall is a judgement, which belongs in a finding. Not Applicable where
	// the policy binds no plan to this requirement.
	"plan-conformance": #Determination @go(PlanConformance)

	// outcome is whether the evidence satisfied the requirement. It is stated for
	// every governed requirement, complete by construction, so that a requirement
	// checked and met is distinguishable from one never checked — which is the
	// property an audit exists to establish. What it means for compliance is a
	// finding's business, not this entry's.
	outcome: #Determination

	// basis is what in the evidence established that outcome. It is for evidence
	// that does not say so itself: an assessment carries its own result, so a
	// citation of one needs no restating, while a PDF or an API response leaves the
	// reader with an artifact and a conclusion and nothing between them.
	basis?: string

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// With no plan bound to the requirement there is nothing to have conformed to.

	if effective."plan-id" == _|_ {
		"plan-conformance": "Not Applicable"
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
