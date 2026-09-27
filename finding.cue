// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="experimental")
package gemara

@go(gemara)

// FindingLifecycle is what became of a finding. Open is outstanding, Resolved
// was dealt with, and Waived means a subject that cannot conform was granted an
// exception, so the finding stands and enforcement stands down. It maps to OCSF's
// finding status_id.
//
// It is independent of #Disposition, which says what an enforcement gate did: a
// blocked merge is Enforced with its finding still Open. The exception is that a
// Tolerated action's findings are Waived. Both relationships are documented rather
// than enforced, as #Result's roll-up precedence is.
#FindingLifecycle: "Open" | "Resolved" | "Waived" @go(-)

// Finding records something an evaluation, an enforcement action or an audit
// found, together with what became of it. A finding is not inherently negative —
// ISO 19011 defines an audit finding as the result of evaluating collected
// evidence against the audit criteria, indicating conformity as readily as
// nonconformity — 
// and it is not a compliance determination: whether a requirement is satisfied is
// decided at the audit, which layers that determination on top of this (see
// #ComplianceFinding). Here a finding carries its lifecycle and its materiality.
//
// The shape is the one OCSF's Compliance Finding expects, so a finding projects
// without rearranging: lifecycle carries status_id, description and severity
// carry finding_info.desc and severity_id, and recommendations carry remediation.
// compliance.status comes from the audit's determination, not from here.
// There is no id here: a finding reported against a named check is identified by
// that check, and a finding that is referenced declares its own id where it lives.
// finding_info.title has no field here on purpose — title is catalog vocabulary
// in Gemara, naming a defined entry, and no other log record carries one. A
// translation takes the title from the requirement the finding concerns.
#Finding: {
	// lifecycle is what became of this finding: still outstanding, dealt with, or
	// waived because a subject that cannot conform was granted an exception. It
	// is deliberately not #Disposition, which says what an enforcement gate did.
	lifecycle?: #FindingLifecycle

	// id identifies this finding where something references it. A finding
	// reported against a named check needs none, so it is optional here and
	// required where it is addressed.
	id?: string

	// requirement names the assessment requirement this finding concerns.
	requirement?: #EntryMapping @go(Requirement,optional=nillable)

	// log names the evaluation log entry this finding came from, when it came
	// from one rather than from an auditor reading evidence directly. There is no
	// plan field beside it: the entry it points at names its own plan, and a second
	// copy here could only disagree with it.
	log?: #EntryMapping @go(Log,optional=nillable)

	// description explains what was found. It is optional because a routine
	// conclusion needs no prose: a requirement recorded Compliant says enough by
	// saying so. An override, or anything a reader would question, explains itself.
	description?: string

	// severity states how material the finding is, on the same scale the risk
	// catalog uses. It is optional here because not every finding rates one: an
	// audit finding that records conformity has no materiality to state. It is
	// required where a finding is negative by construction — anything an audit
	// determines Not Compliant, and every finding in an enforcement
	// justification.
	severity?: #Severity

	// recommendations record the corrective actions for this finding.
	recommendations?: [#Recommendation, ...#Recommendation]
}

// Recommendation provides a corrective action for a finding.
#Recommendation: {
	// id uniquely identifies this recommendation
	id?: string

	// text describes the recommended corrective action
	text: string

	// required indicates whether this recommendation is a mandatory corrective action
	required: *false | bool @gemara(default=false)
}
