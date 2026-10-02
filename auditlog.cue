// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

import "list"

@go(gemara)

// Determination describes the outcome of a performed verification.
#Determination: "Satisfied" | "Not Satisfied" | "Not Applicable" | "Undetermined" @go(-)

// ComplianceStatus is the audit's decision about one control's objective, or
// about a risk. It is the audit's conclusion, which an auditor may reach
// differently from what the assessment observed. Compliance is decided one
// control at a time, so there is no "compliant with exceptions".
#ComplianceStatus: "Compliant" | "Not Compliant" | "Not Applicable" | "Undetermined" @go(-)

// Opinion is the audit's conclusion about the target as a whole, reached after
// weighing all of the findings against what the audit set out to establish.
#Opinion: "Passed" | "Passed with Conditions" | "Failed" | "Undetermined" @go(-)

// AuditLog records an audit, including what it examined the target against, what it
// established for each requirement, what it reports, and what it concludes.
#AuditLog: {
	#Log
	metadata: type: "AuditLog"

	// owner defines the RACI roles responsible for managing the audit
	owner?: #RACI @go(Owner,optional=nillable)

	// opinion is the audit's conclusion in one word, so a consumer of the
	// attestation need not weigh the findings itself.
	opinion: #Opinion

	// summary is that conclusion in prose, including anything the opinion cannot
	// carry: what the objectives established, trends across the findings, and the
	// uncertainty any audit of a moment carries.
	summary: string

	// policy is this audit's criteria: the single versioned artifact the evidence
	// was compared against. It is one policy rather than a list because the policy
	// also defines scope. The audit's completeness claim needs a governed set to be
	// complete against.
	policy: #ArtifactMapping

	// objectives are what this audit set out to establish. They are recorded
	// because the conclusion has to speak to them. An opinion answers a question,
	// and without the question a later reader cannot tell whether the audit
	// achieved what it was for, or judge the scope it settled for.
	objectives: [string, ...string]

	// findings are the audit's decisions about control objectives and risks. A
	// finding about a control decides from the verifications of that control's
	// requirements. It is absent when the audit had nothing to call out.
	findings?: [...#ComplianceFinding] @go(Findings,type=[]*ComplianceFinding)

	// verifications are the mechanical checks of each assessment requirement the
	// policy governs, against the evidence provided. They are complete by
	// construction, with one entry per requirement including the boring passes,
	// and carry verification outcomes rather than compliance decisions.
	verifications: [#VerificationLog, ...#VerificationLog]

	// coverage records, for guidance the policy adopted, the controls this audit
	// relied on to support each guideline and the findings on them. A guideline is
	// not testable, so it is judged through this map rather than verified.
	coverage?: [#GuidelineCoverage, ...#GuidelineCoverage]

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// Every reference-id names a declared mapping-reference or the audit's own
	// id: the audited policy, each verified requirement and its control, each
	// evidence source, and each finding's and guideline's subject.

	// A finding the audit determines Not Compliant states how material it is.
	// Nothing downstream can triage an unrated Not Compliant finding. The rule is
	// an element constraint rather than a conditional inside #ComplianceFinding
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
		_refValidation: "verifications-\(i)":         _refIds & list.Contains(c."effective-requirement".requirement."reference-id")
		_refValidation: "verifications-\(i)-control": _refIds & list.Contains(c."effective-requirement".control."reference-id")
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

		for i, f in findings if f.control != _|_ {
			_refValidation: "findings-\(i)-control": _refIds & list.Contains(f.control."reference-id")
		}

		for i, f in findings if f.response != _|_ {
			_refValidation: "findings-\(i)-response": _refIds & list.Contains(f.response."reference-id")
		}
	}

	// Any determination a reader would question must be decided in a finding on
	// the requirement's control: an outcome that is not Satisfied, and any
	// fidelity check that came out Not Satisfied or Undetermined. Without that
	// finding, the record says the evidence fell short and never says what that
	// means. It may have fallen short of the requirement, or of the way the policy
	// said to obtain it. The silent case this closes is a plan not followed where
	// the requirement was satisfied anyway. An auditor accepted a
	// divergence, and nothing recorded what was accepted or why.
	//
	// Not Applicable is exempt on both, since nothing fell short: no plan was bound,
	// or the requirement did not apply.

	_findingControls: [
		if findings != _|_
		for f in findings if f.control != _|_ {"\(f.control."reference-id")/\(f.control."entry-id")"},
	]

	// Written as an assignment rather than a conditional: an incomplete value is
	// tolerated on the right of an assignment and fatal in an if clause. outcome
	// stays a disjunction until the data makes it concrete, so a conditional
	// here leaves #AuditLog itself unevaluable.

	for i, v in verifications {
		let _checks = [
			if v.fidelity != _|_
			for c in v.fidelity {c.determination},
		]
		_synthesisValidation: "\(i)": true & (
						((v.outcome == "Satisfied" || v.outcome == "Not Applicable") &&
			!list.Contains(_checks, "Not Satisfied") &&
			!list.Contains(_checks, "Undetermined")) ||
			list.Contains(_findingControls, "\(v."effective-requirement".control."reference-id")/\(v."effective-requirement".control."entry-id")"))
	}

	// A plan bound to the requirement obliges a record of how it was followed. The
	// reverse does not hold: entries without a plan-id are the fallback case, checked
	// against the methods the policy accepts generally. The rule lives here rather
	// than on #VerificationLog because "effective-requirement" is a quoted field,
	// reachable only by selector from a loop variable; a field alias would drop it
	// from the generated OpenAPI.

	for i, v in verifications if v."effective-requirement"."plan-id" != _|_ {
		_fidelityValidation: "\(i)": true & (v.fidelity != _|_)
	}

	// A finding names either a control this audit verified, or a risk. Naming a
	// control none of whose requirements has a verification entry would decide
	// about something the audit never checked.

	_verifiedControls: [for v in verifications {"\(v."effective-requirement".control."reference-id")/\(v."effective-requirement".control."entry-id")"}]
	if findings != _|_ {
		for i, f in findings if f.control != _|_ {
			_controlValidation: "\(i)": _verifiedControls & list.Contains("\(f.control."reference-id")/\(f.control."entry-id")")
		}
	}

	// A finding names what it decides about: a control, a risk, or both, since a
	// failed control can leave a residual risk the same finding cites.

	if findings != _|_ {
		for i, f in findings {
			_findingSubjectValidation: "\(i)": true & (f.control != _|_ || f.risk != _|_)
		}
	}

	// Finding ids are unique within the audit, so a coverage entry's citation of
	// a finding is unambiguous.

	if findings != _|_ {
		_uniqueFindingIds: {for i, f in findings {(f.id): i}}
	}

	// A guideline's evidence is only as good as the checking behind it: every
	// control a guideline relies on is one this audit verified, and every finding
	// it cites is a finding on one of those controls. A guideline with no
	// controls behind it states what the audit made of that, since the map then
	// records nothing else. The first two rules need the audit's verifications
	// and findings, so they live here, and the rationale rule sits beside them.

	if coverage != _|_ {
		for i, g in coverage {
			let _keys = [
				if g.controls != _|_
				for c in g.controls {"\(c."reference-id")/\(c."entry-id")"},
			]
			_refValidation: "coverage-\(i)": _refIds & list.Contains(g.guideline."reference-id")
			if g.controls != _|_ {
				for j, c in g.controls {
					_refValidation: "coverage-\(i)-control-\(j)": _refIds & list.Contains(c."reference-id")
				}
			}
			for j, k in _keys {
				_guidelineControlValidation: "\(i)-\(j)": _verifiedControls & list.Contains(k)
			}
			let _eligible = [
				if findings != _|_
				for f in findings
				if f.control != _|_
				if list.Contains(_keys, "\(f.control."reference-id")/\(f.control."entry-id")") {f.id},
			]
			if g.findings != _|_ {
				for j, id in g.findings {
					_guidelineFindingValidation: "\(i)-\(j)": _eligible & list.Contains(id)
				}
			}
			_guidelineRationaleValidation: "\(i)": true & (g.controls != _|_ || g.rationale != _|_)
		}
	}
}

// ComplianceFinding is the audit's decision about one control's objective, or
// about a risk. It does not restate the verification outcomes it decides from;
// it says what they mean for compliance, which takes an auditor's reasoning.
#ComplianceFinding: {
	#Finding

	// status is the audit's decision about the control's objective or the risk.
	status: #ComplianceStatus

	// id uniquely identifies this finding within the audit. It is declared here
	// rather than in the #Finding core because a finding reported against a named
	// check needs no id of its own. This one is reported in its own right.
	id: string

	// response names the enforcement action that produced this finding's current
	// lifecycle: the gate that blocked it, the remediation that resolved it, or the
	// tolerated action whose exception waived it.
	response?: #EntryMapping @go(Response,optional=nillable)

	// risk names the risk this finding concerns, when it concerns one: a residual
	// risk beyond its category's tolerance, for example. A finding names a control,
	// through the #Finding core, or a risk.
	risk?: #EntryMapping @go(Risk,optional=nillable)
}

// GuidelineCoverage is how the audit judged one guideline the policy adopted. A
// guideline is not testable, so nothing verifies it directly: the audit names the
// controls it relied on to support the guideline and cites the findings on them.
// The map is recorded here rather than pointed at in a #MappingDocument because
// an audit is a self-contained record of what it relied on.
#GuidelineCoverage: {
	// guideline names the guideline the policy adopted.
	guideline: #EntryMapping

	// statement is the guideline as the policy adopted it, resolved at audit time
	// so the record says what was judged without fetching the catalog.
	statement: string

	// controls are the controls the audit relied on to support the guideline.
	// Absent means nothing in scope supports it.
	controls?: [#EntryMapping, ...#EntryMapping]

	// findings are the ids of this audit's findings, on those controls, that bear
	// on the guideline.
	findings?: [string, ...string]

	// rationale is how the audit reads the map: what the controls cover of the
	// guideline and what they leave uncovered. It is required where no control
	// supports the guideline.
	rationale?: string
}

// EffectiveRequirement is what a policy required of one assessment requirement,
// as resolved at audit time: the control it tests, the text as it applied, the
// applicability that selected it, the constraints the policy attached, and the
// plan bound to it.
#EffectiveRequirement: {
	// control names the control whose objective this requirement tests. It is
	// recorded rather than derived because an entry id only suggests its control,
	// and the audit decides compliance per control.
	control: #EntryMapping

	// requirement names the assessment requirement this effective requirement
	// builds on.
	requirement: #EntryMapping

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
// requirement, as it applied.
#EffectiveConstraint: {
	// id identifies this constraint within the policy that attached it.
	id: string

	// text is the constraint as it applied to the requirement.
	text: string
}

// VerificationLog records the mechanical checks for one governed requirement:
// what was examined, whether the evidence was obtained the way the policy
// required, and whether it satisfied the requirement. Findings decide from these
// entries, which establish how far the evidence can be trusted and what it shows.
#VerificationLog: {
	// effective-requirement is what the policy required, resolved at audit time,
	// so the record says what was audited without fetching the policy or catalog.
	"effective-requirement": #EffectiveRequirement

	// outcome is whether the evidence satisfied the requirement.
	outcome: #Determination

	// rationale is what in the evidence established these determinations, for
	// evidence that does not say so itself. An assessment carries its own result,
	// so a citation of one needs no restating.
	rationale?: string

	// assessments are the assessments this audit read, if any.
	assessments?: [#EvidenceMapping, ...#EvidenceMapping]

	// evidence records the data sources that support this outcome.
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)

	// fidelity records whether the evidence was obtained the way the policy
	// prescribed, as one entry per thing the audit checked. It concerns how the
	// evidence was produced; whether the requirement was satisfied is outcome.
	fidelity?: [#FidelityCheck, ...#FidelityCheck]

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// Each evidence entry carries a payload, a source, or both.

	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]
}

// FidelityCheck is one check of how faithfully evidence was obtained to what the
// policy prescribed, for example: that a method the plan accepts was used, that
// the executor matches the one the plan names, that a run fell inside the
// requirement's cadence, that parameter values were within the accepted ones.
#FidelityCheck: {
	// method-id names the accepted method this entry concerns, where it concerns one:
	// one of the plan's where a plan is bound, or one of the policy's fallback methods
	// where none is.
	"method-id"?: string @go(MethodId)

	// determination is what the check came to.
	determination: #Determination

	// required is what the policy prescribed, quoted or derived from it.
	required?: string

	// evidenced is what the evidence showed instead.
	//
	// Both are for divergences. A check that passed needs neither: the method and
	// the determination already say the prescribed thing was done. Restating the
	// prescription on every passing check makes a complete record unreadable. In
	// an audit of a real baseline the passing checks outnumber the failing ones
	// fifty to one. Where the two differ, state both, because the pair is what a
	// determination alone throws away: required every 30 days, evidenced every 90.
	evidenced?: string
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
	source?: #EvidenceMapping @go(Source,optional=nillable)

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
