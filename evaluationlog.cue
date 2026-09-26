// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

import "list"

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

	_refIds: _ // computed on #Log; named here only so this block can reference it

	// A ControlEvaluation carries no id of its own, by design: it is addressed by
	// the control it evaluated, which control (an EntryMapping) already names.
	// That makes (this log's metadata.id, control.reference-id, control.entry-id)
	// a natural composite key, and AssessmentFinding.log resolves against it, but
	// only if a log evaluates each control at most once.

	_uniqueEvaluatedControls: {
		for i, e in evaluations {"\(e.control."reference-id")/\(e.control."entry-id")": i}
	}

	// Every reference-id in this log names one of its declared mapping-references,
	// or the log's own metadata.id.

	for i, e in evaluations {
		_refValidation: "evaluations-\(i)": _refIds & list.Contains(e.control."reference-id")
		for j, a in e."assessment-logs" {
			_refValidation: "evaluations-\(i)-\(j)": _refIds & list.Contains(a.requirement."reference-id")
		}
	}

	// An AssessmentLog is one method's execution against one requirement under
	// one plan, so (this log, requirement, plan-id, method-id) is its key. An
	// assessment with no plan has no method to select and falls back to the
	// requirement alone.

	for i, e in evaluations {
		_uniqueAssessments: "\(i)": {
			for j, a in e."assessment-logs" {
				let _plan = [if a.plan != _|_ {a.plan."entry-id"}, "-"][0]
				let _method = [if a.execution != _|_ {a.execution."method-id"}, "-"][0]
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
	// assessment-logs records each assessment procedure executed against this
	// control's requirements.
	"assessment-logs": [#AssessmentLog, ...#AssessmentLog] @gemara(projectable=false) @go(AssessmentLogs,type=[]*AssessmentLog)

	// ---- Validation --------------------------------------------------------

	// Each assessment's requirement belongs to the evaluated control's catalog;
	// an assessment that omits the reference inherits the control's.

	"assessment-logs": [...{
		requirement: "reference-id": (control."reference-id")
	}]

	// Assessments that actually executed record a start time.

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

// AssessmentLog contains the results of executing a single assessment procedure for a control requirement.
#AssessmentLog: {
	// Requirement should map to the assessment requirement for this assessment.
	requirement: #EntryMapping
	// plan names the policy assessment plan being executed, in the policy it comes
	// from. It is a mapping rather than a bare plan id because an evaluation log
	// has to be readable on its own: most are never audited, so the log cannot
	// lean on a downstream audit to say which policy governed the run. An audit
	// compresses the same pointer to effective.plan-id, which is safe there only
	// because the audit pins its policy once at the document level.
	plan?: #EntryMapping @go(Plan,optional=nillable)
	// Description provides a summary of the assessment procedure.
	description: string

	// execution records what this run used under its plan. A cited plan obliges
	// the record; without a plan there is nothing to record against.
	execution?: #ExecutionFacts

	// Result is the aggregate outcome of the assessment procedure, by the precedence documented on #Result.
	result: #Result
	// Message provides additional context about the assessment result.
	message: string
	// Applicability is elevated from the Layer 2 Assessment Requirement to aid in execution and reporting.
	applicability: [string, ...string] @go(Applicability,type=[]string)
	// Steps are sequential actions taken as part of the assessment, which may halt the assessment if a failure occurs.
	steps: [#AssessmentStep, ...#AssessmentStep]
	// Steps-executed is the number of steps that were executed as part of the assessment.
	"steps-executed"?: int & >=0 & <=len(steps) @go(StepsExecuted)
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

	// ---- Validation --------------------------------------------------------

	// A cited plan obliges a record of how it was followed; without a plan there
	// is nothing to record against.

	if plan != _|_ {execution: #ExecutionFacts}

	if plan == _|_ {execution?: _|_}

	// Each evidence entry carries a payload, a source, or both.

	evidence?: [#_EvidenceStrict, ...#_EvidenceStrict]
}

#AssessmentStep: string @go(-)

// Result is an assessment outcome. Aggregation at every rollup (assessment,
// control evaluation, log) follows the precedence
// Failed > Unknown > Needs Review > Passed > Not Applicable, and Not Run never
// overwrites another value.
#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)

// ExecutionFacts records what a run actually used under its plan: which of the
// plan's methods ran, who ran it when that differs from the log's author, and
// the parameter values used. A log cites a plan to say what was supposed to
// happen; these are what did.
#ExecutionFacts: {
	// method-id names the accepted method from the cited plan that this
	// assessment executed. Method ids are unique across a policy, so no further
	// qualifier is needed.
	"method-id": string @go(MethodId)

	// executor is the actor that performed this assessment, recorded only when
	// it differs from the log's metadata.author.
	executor?: #Actor

	// parameters are the selected values for execution correlated to the plan's parameters
	parameters?: [#ParameterValue, ...#ParameterValue]
}

// ParameterValue records the value an execution used for a plan parameter.
// The id names a #Parameter declared by the plan that was executed.
#ParameterValue: {
	id:    string
	value: string
}
