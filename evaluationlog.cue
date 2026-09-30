// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// EvaluationLog contains the results of evaluating a set of Layer 2 controls.
#EvaluationLog: {
	#Log
	metadata: type: "EvaluationLog"
	// result is the aggregate outcome across all evaluations in this log
	result: #Result
	evaluations: [#ControlEvaluation, ...#ControlEvaluation] @go(Evaluations,type=[]*ControlEvaluation)
	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// An AssessmentLog is one method's execution against one requirement under one
	// plan, so (this control evaluation, requirement, plan-id, method-id) is its key.
	// This is what lets a log hold several results for one requirement — a plan
	// requiring an automated probe and a manual review produces two — and what stops
	// it holding two that nothing tells apart. An assessment citing no plan has no
	// method to select, so the requirement alone must be unique within its control
	// evaluation: attribution by executor does not make two such records distinct.

	for i, e in evaluations {
		_uniqueAssessments: "\(i)": {
			for j, a in e."assessment-logs" {
				let _plan = [if a.plan != _|_ {a.plan."entry-id"}, "-"][0]
				let _method = [if a.execution != _|_ if a.execution["method-id"] != _|_ {a.execution."method-id"}, "-"][0]
				"\(a.requirement."entry-id")/\(_plan)/\(_method)": j
			}
		}
	}

}

// ControlEvaluation contains the results of evaluating a single Layer 5 control.
#ControlEvaluation: {
	name:    string
	result:  #Result
	message: string
	control: #EntryMapping
	"assessment-logs": [#AssessmentLog, ...#AssessmentLog] @gemara(projectable=false) @go(AssessmentLogs,type=[]*AssessmentLog)
	// Enforce that control reference and the assessments' references match
	// This formulation uses the control's reference if the assessment doesn't include a reference
	"assessment-logs": [...{
		requirement: "reference-id": (control."reference-id")
	}]
	// Require start timestamp on assessments that actually executed
	"assessment-logs": [#_AssessmentLogStrict, ...#_AssessmentLogStrict]
}

// _AssessmentLogStrict layers the "start required unless unexecuted" rule on top of #AssessmentLog
#_AssessmentLogStrict: {
	@go(-)
} & #AssessmentLog & {
	result: #Result
	if result != "Not Run" && result != "Unknown" && result != "Not Applicable" {
		start: #Datetime
	}
}

// ExecutionConfiguration records how an assessment was configured: which method it
// ran, and the values it ran with. The name is qualified on purpose — in this domain
// a bare "configuration" reads as the target's, since a scanner's evidence is so
// often the repository's own settings, and this is the configuration of the run.
//
// What the assessment *is* — its result, timing, steps, and the executor that
// produced it — stays at the top level of #AssessmentLog, where it is meaningful
// whether or not a policy exists. Who performed the execution is executor, not part
// of its configuration.
//
// It is one field rather than several so that a stable type does not gain a permanent
// top-level field per configuration input; further inputs land here additively, where
// a top-level field could never be moved.
//
// Not everything here need come from a policy. A policy-aware producer records the
// method the policy prescribed and the values the plan declared; a producer that knows
// of no policy still has a configuration, and this is where it goes. Only the parts a
// policy prescribed are comparable against it, which is a property of those fields
// rather than of this container: an audit checks method-id and declared parameter
// values, and treats anything else as evidence about the run.
#ExecutionConfiguration: {
	// method-id names the accepted method this assessment executed, where the policy
	// prescribed one: a method of the cited plan, or one of the policy's own
	// evaluation-methods, which are the fallback where no plan is bound. Method ids
	// are unique across a policy, so the id resolves without saying which of the two
	// declared it — unlike #ActionResult.method, which carries a full mapping because
	// an enforcement action cites no plan.
	"method-id"?: string @go(MethodId)
}

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: #EntryMapping
	// Plan maps to the policy assessment plan being executed.
	plan?: #EntryMapping @go(Plan,optional=nillable)

	// execution records how this assessment was configured. A cited plan obliges a
	// method: the plan lists the methods it accepts, so a record of running under it
	// has to say which one ran.
	execution?: #ExecutionConfiguration @go(Execution,optional=nillable)
	// executor is the actor that performed this assessment: the tool instance,
	// service or person that produced this result, not the author of the log. A
	// multi-source log is assembled by one orchestrator — metadata.author — from
	// results several executors produced, so attribution per result cannot be derived
	// from the document. It is recorded only where it differs from metadata.author,
	// since repeating the author on every assessment would be a second copy of one
	// fact.
	//
	// Where the cited plan's accepted method also names an executor, this records
	// what actually ran and that records what was meant to. The two disagreeing is
	// not an error: an assessment performed by an instance the plan does not accept
	// is a real event, and a schema that rejected the record would leave a producer
	// choosing between lying and omitting the attribution this field exists for. An
	// audit reports the mismatch as a plan-conformance failure, comparing by
	// authoritative identifier rather than by display name. Neither document
	// overrides the other, and no consumer should silently reconcile them.
	//
	// For that reason, where the plan's method names an executor, record this field
	// explicitly rather than leaning on the omission convention: an omission asserts
	// that the log's author ran the assessment, which is itself a disagreement with
	// the plan, and a silent one.
	//
	// Attribution does not deconflict: two results for one requirement are told
	// apart by the plan's methods, per the identity rule below, not by who produced
	// them.
	executor?: #Actor @go(Executor,optional=nillable)

	// Description provides a summary of the assessment procedure.
	description: string
	// Result is the overall outcome of the assessment procedure, matching the result of the last step that was run.
	result: #Result
	// Message provides additional context about the assessment result.
	message: string
	// Applicability is elevated from the Layer 2 Assessment Requirement to aid in execution and reporting.
	applicability: [string, ...string] @go(Applicability,type=[]string)
	// Steps are sequential actions taken as part of the assessment, which may halt the assessment if a failure occurs.
	steps: [#AssessmentStep, ...#AssessmentStep]
	// Steps-executed is the number of steps that were executed as part of the assessment.
	"steps-executed"?: int @go(StepsExecuted)
	// Start is the timestamp when the assessment began.
	// Assessments that never executed have no start time to record.
	start?: #Datetime

	// End is the timestamp when the assessment concluded.
	end?: #Datetime
	// Recommendation provides guidance on how to address a failed assessment.
	recommendation?: string
	// ConfidenceLevel indicates the evaluator's confidence level in this specific assessment result.
	"confidence-level"?: #ConfidenceLevel @go(ConfidenceLevel)
	// Evidence records the raw data cited to support this assessment's opinion.
	evidence?: [#Evidence, ...#Evidence] @go(Evidence)
	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// A cited plan obliges a method: the plan lists the methods it accepts, and a
	// record of running under it that does not say which one ran leaves two results
	// against the same plan indistinguishable. Stated on the field it requires —
	// forbidding a field instead puts bottom in that field's type, and the Go
	// projection degrades it to an unusable `any`.
	//
	// There is no rule the other way. Settings without a plan is the ordinary case
	// for a producer running the policy's fallback method, or for one that knows of
	// no policy at all.

	if plan != _|_ {execution: "method-id": string}

}

#AssessmentStep: string @go(-)

#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
