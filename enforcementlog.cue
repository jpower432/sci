// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

@go(gemara)

// EnforcementLog records actions taken in response to noncompliance findings from Layer 5 evaluations.
#EnforcementLog: {
	#Log
	metadata: type: "EnforcementLog"
	// disposition is the aggregate across this log's actions, by the precedence
	// documented on #Disposition.
	disposition: #Disposition
	// actions is the list of enforcement actions performed
	actions: [#ActionResult, ...#ActionResult] @gemara(projectable=false) @go(Actions,type=[]*ActionResult)

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// Clear means the gate ran and found nothing to act on, so it has nothing to
	// justify. Every other disposition carries a justification that justifies
	// something: findings, or the exceptions that stood the enforcement down. An
	// empty justification would satisfy #Justification, whose every field is
	// optional, while saying nothing.

	actions: [...#_ActionResultStrict]

	// Action ids are unique within the log, so an action can be referenced
	// unambiguously.

	_uniqueActionIds: {for i, a in actions {(a.id): i}}

	// Finding ids must be unique across the whole log, not just within one
	// action's justification: a reference to a finding id would otherwise be
	// ambiguous between actions. The value names the action position, so a
	// collision names both.

	_uniqueFindingIds: {
		for i, a in actions if a.justification != _|_ if a.justification.findings != _|_ for j, f in a.justification.findings {
			(f.id): "actions[\(i)].justification.findings[\(j)]"
		}
	}
}

// ActionResult captures a performed enforcement action.
#ActionResult: {
	// id allows this entry to be referenced by other elements
	id: string

	// disposition is the enforcement action taken
	disposition: #Disposition @go(Disposition)

	// method references the specific AcceptedMethod entry within the Policy being enforced
	method: #EntryMapping @go(Method)

	// message provides additional context about the action
	message?: string @go(Message,type=*string)

	// start is the timestamp when the enforcement action began
	start: #Datetime

	// end is the timestamp when the enforcement action concluded
	end?: #Datetime

	// steps references the code paths or addresses that carried out this enforcement action
	steps: [#EnforcementStep, ...#EnforcementStep]

	// justification links the action to the findings that prompted it and any
	// applicable exceptions. A Clear action has nothing to justify, so it is
	// absent there and required everywhere else.
	justification?: #Justification @go(Justification,optional=nillable)
}

// EnforcementStep is a reference to the code that performed an enforcement action
#EnforcementStep: string @go(-)

// Justification provides the assessment data and exception references that justify an enforcement action.
#Justification: {
	// findings links the action to the findings that justify it. Each carries an
	// id, so an action or a later log can reference one unambiguously; a log
	// reference, because an enforcement action reacts to an evaluation, so the
	// finding it acts on came from one; and a severity, because a finding a gate
	// acted on is negative by nature. Unlike an audit finding, which may record
	// conformity as readily as nonconformity, there is nothing to enforce against
	// good news, and a waived Critical is not a waived Low. This is the grading of
	// nonconformities that certification bodies work to under ISO/IEC 17021-1,
	// which #Severity carries.
	findings?: [#Finding & {id: string, log: #EntryMapping, severity: #Severity}, ...#Finding & {id: string, log: #EntryMapping, severity: #Severity}] @go(Findings)

	// exceptions references the approved exceptions that authorise this action:
	// a subject that cannot conform has been granted one, so evaluation still
	// fails and enforcement stands down. Gemara does not model an exception, so
	// these name external artifacts through the log's mapping-references.
	exceptions?: [#ArtifactMapping, ...#ArtifactMapping] @go(Exceptions)
}

// Disposition is what an enforcement gate did. It rolls up from an action to the
// log by the precedence Enforced > Tolerated > Undetermined > Clear: a log where
// anything was enforced is Enforced, and Clear only survives when every action
// was. The strongest action dominates because a disposition records what the
// enforcement layer did; how much that should worry anyone is the audit's
// question, answered there by determinations, severities and lifecycles.
//
// A disposition is not a finding's lifecycle. Blocking a merge is Enforced while
// the finding it blocked stays Open, since the gate prevented harm without fixing
// anything; a remediation is also Enforced and does leave its finding Resolved.
// The one correspondence that always holds: a Tolerated action's findings are
// Waived, because accepting something without action is what waiving it means.
#Disposition:
	// Enforcement outcome could not be determined.
	"Undetermined" |
	// Findings existed and actions were taken.
	"Enforced" |
	// Findings existed but were accepted without action.
	"Tolerated" |
	// No findings, nothing to act on.
	"Clear" @go(-)

// #_ActionResultStrict carries the justification rules on #ActionResult. It is
// hidden (@go(-), never projected), so #ActionResult stays shape and
// documentation only, and the rules do not leak into the OpenAPI projection as a
// stub property on the actions list. disposition is redeclared here so the
// conditions resolve to the action rather than to the log's aggregate.
#_ActionResultStrict: {
	@go(-)
} & #ActionResult & {
	disposition: #Disposition

	if disposition == "Clear" {
		justification?: error("a Clear action found nothing to act on, so it has nothing to justify")
	}
	if disposition != "Clear" {
		justification: #Justification

		if justification.findings == _|_ if justification.exceptions == _|_ {
			justification: findings: error("an action that is not Clear justifies itself: name the findings it acted on, or the exceptions that stood it down")
		}
	}
}
