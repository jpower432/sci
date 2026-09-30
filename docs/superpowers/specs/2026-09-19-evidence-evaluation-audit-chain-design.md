# Evidence, evaluation and audit: completing the define → record → check chain

Date: 2026-09-19
Branch: `fix/cue2openapi`
Schema baseline: `6e463c0` (v0.17.0-dev)
Source: Gemara Field Lineage audit

**Boundary: the four types.** This redesign covers `#AssessmentPlan`, `#AssessmentLog`,
`#Evidence` and `#AuditLog`, and the definitions they own. Audit threads outside those
types — the resolution pipeline, the mapping primitives work stream, the L1/L2 catalog
rules, the quality-control phases — are out, however adjacent. They are a separate change.

---

## 0. As built (2026-09-26)

Built and committed on `audit-flow-redesign`, ten commits, all gates green, PR pending.
Sections 1–3 still hold. Section 4 records the design **as decided** and has drifted in
names and in two decisions; this section supersedes it. Section 8's sequencing is history.
The blow-by-blow divergence, with reasons, is in
`.superpowers/sdd/2026-09-19-evidence-evaluation-audit-chain/progress.md`.

### Renames

| Designed | Built |
|---|---|
| `#EvidenceExpectation` | `#EvidenceRequirements` |
| `#RequirementCoverage`, field `coverage` | `#VerificationLog`, field `verifications` |
| `#CheckOutcome` | gone — a check is a bare `#Determination` |
| `#AuditResult` | `#ComplianceFinding` |
| `assessment-present` | `evidence-present` |
| `mapping_primitives.cue` | `primitives.cue` |
| — | `finding.cue`, new: the `#Finding` core the three logs share |

### Decisions that changed

**D2 is not what was built.** The design had the EvaluationLog carry the effective
requirement set. The built chain puts it in the audit: `#VerificationLog.effective` holds
the resolved text, applicability, constraints and `plan-id`, per requirement. The reason is
D1's own logic — the audit is the point-in-time attestation, so the resolution it checked
belongs in it, and an evaluation log that carried the set would be asserting something it
has no standing to assert.

**D3's split survived; its vocabulary did not.** `verifications` is mechanical and
complete, `findings` is selective and reasons — but each verification entry now also states
an `outcome`, because a reader who has to derive per-requirement compliance from the checks
or cross-reference the findings is doing the recomputation this layer exists to prevent.

### Decisions added during implementation

| # | Decision |
|---|---|
| D9 | Four vocabularies, one per scope: `#Result` (did a procedure pass), `#Determination` (does a named check's condition hold), `#ComplianceStatus` (does the target comply with one requirement), `#Opinion` (did the target pass the audit) |
| D10 | `#Opinion` exists and is stated, not derived — ISO 19011's audit conclusion, since weighing exceptions for materiality is a judgement no precedence computes |
| D11 | `findings` is optional. A clean audit has nothing to call out, and the required list made that unrepresentable |
| D12 | Cardinality propagates along the chain: required in the plan ⇒ named in the record ⇒ verified in the log. This is what made `allowed`, `used` and `cadence-met` required per method |
| D13 | A digest is one value, not a set. The rationale, and the in-toto trigger that would reopen it, are on `#Digest` |
| D14 | `download-url` requires `digest` — the enforceable half of ISO 19011's verifiable evidence |
| D15 | Severity is required where a finding is negative by construction: anything `Not Compliant`, and every finding in an enforcement justification |
| D16 | `evidence-present` may be omitted only for a `Not Applicable` outcome. Found by authoring the full baseline: it was ceremony on 16 of 56 entries |

### Rules that landed in CUE beyond the design

`effective.plan-id` ⇔ `plan` in both directions; a plan verification enumerating at least
one method; an entry claiming coverage naming the evidence or assessments behind it; unique
plan `requirement` across a policy; a plan's requirement reference being declared; catalog
`extends`/`imports` membership; and on the enforcement side, a justification that justifies
something — `#Justification` has no required field of its own, so requiring the struct was
vacuous.

### Two systemic findings, both worth knowing before touching these rules

**A conditional on a required enum leaves the definition unevaluable.** `cue vet` passes on
concrete data while the Go harness fails, and the error names the wrong field. Two homes for
such a rule: a hidden `#_XStrict` variant (counted by `TestGenerateRawProducesSchemas`, not
projected), or an inline element constraint (which leaks a stub property into the published
items schema). Prefer restating the rule as a contrapositive over **optional** fields, which
is decidable and needs neither.

**`x == _|_` is true of an invalid value as well as an absent one.** Three rules cascade a
second, misleading error when a value is malformed rather than missing, and the misleading
one can sort first. There is no fix in CUE. The mitigation is fixture discipline: a negative
fixture must satisfy every rule but the one under test.

### Where the prose went

`docs/tool-spec.md` — the cross-document, procedural and judgement rules, with the ISO 19011
and ISO/IEC 17021-1 citations. `examples/osps-level-2` — the complete chain at the scale of
a real baseline: policy, scan, gate log, audit. `test/conformance_test.go` — the tool spec's
clock-free rules, checked against that chain, each one proven to fail against a broken copy.

---

## 1. Problem

Gemara expresses compliance as three stations: a Policy **defines** what is required
(`#AssessmentPlan`), an EvaluationLog **records** what was done (`#AssessmentLog`), and an
AuditLog **checks** the second against the first. The audit found the same break at the
same place repeatedly: the define station is rich, the record station is thin, and the
check station does not exist as a schema concept at all.

On the current schema:

- `#AssessmentPlan` (policy.cue:108-115) declares `frequency` as free prose, an
  `evidence-requirements` string with no consumer, `parameters` no log echoes, and
  `evaluation-methods` whose `required` flag has no stated meaning.
- `#AssessmentLog` (evaluationlog.cue:45-75) can cite a `plan` (evaluationlog.cue:49) and
  records nothing about executing it — not which method ran, who ran it, with what
  parameter values, or what triggered the run (21, 22c, 23).
- `AuditLog.criteria` (auditlog.cue:23) is `[#ArtifactMapping, ...]`: any artifact is a
  valid yardstick. `#AuditResult` is narrative opinion. No field anywhere lets an audit
  state that it verified anything.
- `#Evidence` (auditlog.cue:82-100) lives in an experimental file while stable
  `evaluationlog.cue` consumes it (1h); its `source.reference-id` resolves against nothing
  (29c); its digest has no defined use (29f).

Nearly every thread in this area ends "tool spec: the audit confirms…", and there is
nothing for a tool to write the confirmation into.

The consequence that matters most: an audit cannot say **"two requirements were never
assessed."** It can only report on what a log chose to include, so *checked and passed* and
*never checked* are indistinguishable — the one property an audit exists to establish.

## 2. Decisions

| # | Decision | Rejected |
|---|---|---|
| D1 | An audit's criteria is exactly one versioned **Policy**. Auditing a raw baseline means authoring a Policy that imports it. | A list of arbitrary artifacts; a list with one entry required to be a Policy |
| D2 | The **EvaluationLog records the effective requirement set** it ran against; the audit checks that record against the Policy. | The audit re-deriving the set; a new resolved-policy artifact |
| D3 | The audit produces **coverage** (per requirement, complete, mechanical) separately from **results** (selective, narrative). | Expressing checks as `#AuditResult` entries — thread 16's precedent |
| D4 | Coverage checks are **dedicated named fields**, not an open enum of check names. | `{check, outcome}` pairs over an open or closed enum |
| D5 | `#Evidence` moves to `evidence.cue`, tagged stable, from `feat/evidence-digest-semantics-490`. | Leaving it in experimental `auditlog.cue` (1h) |
| D6 | The four types change together on one branch, with the prerequisites their pointers need. | Landing evidence, plan, log and audit separately |
| D7 | The redesign stays within **v1**. | Treating the `criteria` replacement as breaking |
| D8 | An assessment is **one method's execution** against one requirement under one plan. Its identity is the 4-tuple **(evaluation log, requirement, `plan-id`, `method-id`)**. | Minting ids; the (log, requirement) pair and the (log, requirement, plan) triple, both of which make a plan's second required method unrepresentable |

Rationale for D4, which took the most discussion: absence must be meaningful and readable.
With `{check, outcome}` pairs, "frequency was never checked" and "frequency passed" are
both just list contents, and no consumer has a fixed vocabulary to distinguish them. Named
fields make a missing check visible as a missing field. An open enum would also reintroduce
the exact pattern this work removes — a free-form identifier with no consumer, the shape of
`frequency: string`, `evidence-requirements: string`, and evidence types matched by value.

**The identity test (refines the audit's "derive identity, don't mint it").** Mint an id
only for something that has nothing to derive one from. The audit's phrasing — mint only
for what is *referenced today* — is a snapshot of current consumers, and a reference added
later finds no key to use. Across this design the test decides every case the same way: an
assessment derives from (log, requirement, `plan-id`, `method-id`); a coverage entry from
its requirement; an effective requirement from its pointer. `#Evidence` is the single type
with nothing to derive from — an inline payload has no address, and (source, timestamp)
collides when two items come from one system at once — so its `id` stays **required**, as
today. That settles 29b.

Rationale for D2 under this narrower boundary: a Policy never lists what it requires, and
six of the eight resolution steps are undefined (R2). The snapshot is what lets the audit
have a denominator **without** this change having to define resolution. Resolver
disagreement becomes a reportable finding against a written-down claim instead of a silent
mismatch.

### Reversible calls

- **One evidence expectation per plan**, not a list (29d).
- **`frequency-days` is required**, as `frequency` is today.

## 3. Scope

**In — threads that touch the four types:** 14c, 16, 17 (audit half), 21, 22a, 22b, 22c,
23, 24 (doc only), 26a, 26b, 28 (superseded), EV, 29a, 29b, 29c, 29d, 29e, 29f.

**Prerequisites, because the four types' pointers are worthless without them:**

- **1a + 7b** — every `reference-id` names a declared mapping-reference, or the document's
  own `metadata.id`. Today only uniqueness is checked (metadata.cue:62-63). This is what
  makes `AuditLog.policy`, `Evidence.source` (29c) and coverage's pointers real. Lands as a
  prerequisite commit; if it needs to go in on its own, it can.
- **1b** — `Metadata.version` required. D1 calls for a *versioned* Policy reference, and
  `#MappingReference.version` is already required (mapping_inline.cue:18) while the artifact
  it pins need not declare a version.
- **`worktree-fix-entrymapping-ids`** (#499) — `#EntryMapping` targets get real identity
  with no new stored ids. D8 builds on it rather than re-deriving it.

**Out.** The resolution pipeline (R1–R5, 12b, 19a, 19c, 19d) and the Policy import fields
it would need; the mapping primitives work stream beyond 1a/7b; L1/L2 catalog rules (3, 6,
7a, 13, 15, L1a, L1b, 4a); MappingDocument (M1–M5); enforcement disposition (27a); 1h's
full 19-definition consolidation (only `#Evidence`'s own move is in); the audit's
quality-control phases and their new CI gates.

**Other PRs.** #451 is closing; its content is absorbed. **#482 and #493 are not merging**
— this lands first and they rebase. D8 leaves room for #482's multi-source evaluation
without adopting its schema.

## 4. Design

### 4.1 Define — `#AssessmentPlan` (policy.cue, experimental)

```cue
#AssessmentPlan: {
    id:               string
    "requirement-id": #EntryMapping                       // was: bare string (22b)
    "frequency-days": int                                 // was: frequency, free string (23)
    "evaluation-methods": [#AcceptedMethod & {type: #EvaluationMethodType}, ...]
    "evidence-expectation"?: #EvidenceExpectation         // was: evidence-requirements (29d)
    parameters?: [#Parameter, ...#Parameter]
}

#EvidenceExpectation: {
    description:      string   // what this requirement needs, in prose
    "valid-for-days": int      // how long a collected item stays usable
}
```

- `requirement-id` as `#EntryMapping` names the catalog as well as the requirement. A bare
  id is a guess once a policy imports two catalogs, and it is coverage's join key.
- `frequency-days` puts the unit in the name, matching `valid-for-days`. A free string
  cannot be compared against timestamps.
- **Cadence belongs to the requirement, not to a method or a plan instance.** How often
  something is reviewed is a property of the thing being reviewed. This is enforced rather
  than asserted: **`requirement-id` is unique across a policy's `assessment-plans`**, an
  intra-document rule CUE holds, so one requirement has exactly one plan and therefore
  exactly one cadence with nothing to disagree with. Multi-source evaluation does not need
  a second plan — D8's 4-tuple already supports several methods inside one plan. If #482
  later needs the rule relaxed, relaxing a constraint is non-breaking, and `plan-id` stays
  in the identity key so it survives that relaxation.
- **Validity is a different clock from cadence.** `valid-for-days` measures how long a
  collected item stays probative; `frequency-days` measures how often the requirement is
  re-examined. Neither derives from the other — evidence can outlive several review cycles
  (an annual penetration test), or expire between them (a scan result that is stale in a
  week). They are declared separately and checked separately.
- `#EvidenceExpectation` carries only what a checker needs, and declares no evidence
  *types*: an evidence requirement belongs to a policy and a requirement, and does not
  reach down to what an instance is (29a). `#EvidenceTypeDefinition` does not return.
- Doc comments: `#AcceptedMethod.required` means an assessment run under this plan MUST use
  this method (21). `#Constraint` drops the false "Layer 5/6 tracking" claim and states
  that a constraint is part of the requirement produced by resolution, so an assessment run
  under a plan is run against the constrained requirement (24). Doc only — there is no
  separate constraint check anywhere in this design, because a constraint is folded into
  the requirement set rather than tracked alongside it.

### 4.2 Record — `#AssessmentLog` (evaluationlog.cue, stable; additive only)

```cue
#AssessmentLog: {
    // ...existing fields...
    execution?: #ExecutionFacts     // what this run actually used, under the plan
}

#ExecutionFacts: {
    "method-id":  string                                // which of the plan's methods ran (21, 22c)
    executor?:    #Actor                                // only when it differs from the run's (#495)
    parameters?:  [#ParameterValue, ...#ParameterValue] // {id, value} (23)
}

if plan != _|_ { execution: #ExecutionFacts }   // citing a plan obliges the record
if plan == _|_ { execution?: _|_ }

#EvaluationLog: {
    #Log
    policy?: #ArtifactMapping
    "effective-requirements"?: [#EffectiveRequirement, ...#EffectiveRequirement]
    if policy != _|_ { "effective-requirements": [#EffectiveRequirement, ...] }
    // ...existing result, evaluations...
}

#EffectiveRequirement: {
    requirement: #EntryMapping        // the AR, in its catalog
    "plan-id"?:  string               // part of assessment identity (D8)
}
```

No fixture cites a plan today (22a), so the conditional tightening bites only on new data.

**The recorded facts are grouped, not scattered.** `plan` says what was supposed to happen;
`execution` says what did. Three reasons over three sibling fields:

- **It collapses the rules.** Flat needed three — `method-id` required under a plan,
  `executor` required under a plan, `parameters` forbidden without one. Grouped there is
  one: `plan` present ⇔ `execution` present. A parameter value cannot appear without a plan
  because the struct holding it cannot. A rule that no longer has to be stated is one that
  cannot be got wrong.
- **It keeps a stable type from accreting.** Every field added to `evaluationlog.cue` is
  permanent for v1; one optional field is a smaller permanent surface than three.
- **It matches the chain.** Definition and record sit side by side, and the audit's checks
  read one addressable thing instead of gathering fragments.

`start`/`end`, `steps`, `steps-executed` and `evidence` stay outside it — they exist whether
or not there is a plan, so folding them in would make them conditionally required. `plan`
stays put too: it is already on the stable type, so moving it inside would be a rename
needing add-then-deprecate for no gain.

Three things deliberately *not* added, because they already exist:

- **No `run` on `#AssessmentLog`.** The 488 branch puts `#Run` on `#Log`
  (collections.cue:47), one run per log, and `AssessmentLog.start` already exists. The
  cadence check reads the log's `run.trigger` plus each assessment's `start`.
- **No applicability on `#EffectiveRequirement`.** `#AssessmentLog.applicability` already
  records the selection a run used (evaluationlog.cue:57); 14c is a doc clarification of
  that field, not a new one. A second copy in the snapshot could only disagree with it.
- **`execution.executor` is recorded only when it differs from the run's**
  (`#Run.executor`). Otherwise every assessment repeats the same actor.

When the audit compares a log's `executor` with the plan's, it compares **authoritative
identifiers** — purl, ARN, SPIFFE — not the internal short-name `id`, per "short names
inside, authoritative ids across" (4b).

**The snapshot is a receipt, not a source of truth.** Identity and resolution outcome only
— no requirement text, no copied constraint prose, no inlined plans. A reader who wants the
words follows the pointer to the artifact at its pinned version, which is what pinning is
for. This stops the snapshot becoming a second copy of the catalog that can drift from it.
Coverage, plan conformance, cadence and evidence validity join on identity alone.

**Identity (D8).** An assessment is one method's execution against one requirement under
one plan, identified by (evaluation log, requirement, `plan-id`, `method-id`) and unique on
the 4-tuple. Nothing is minted — all four parts already exist.

This follows from what running a plan means: the plan declares a list of
`evaluation-methods`, a run **selects one of them**, and that selection is what the log
records. It is why the tightening above makes `execution` — and so `method-id` — required
whenever `plan` is cited. A plan may declare two `required` methods — a manual Intent review and an automated
Behavioral probe — and each produces its own assessment with its own result, steps, timing
and evidence. `ControlEvaluation.result` rolls them up by the precedence 26a makes
canonical (Failed > Unknown > Needs Review > Passed > Not Applicable, `Not Run` ignored).

The narrower keys were both tried and rejected: (log, requirement) and (log, requirement,
`plan-id`) each permit only one assessment per plan, which makes a plan's second required
method unrepresentable and would make the audit report a method that ran and passed as not
used.

An assessment with no plan has no method to select, so it falls back to (evaluation log,
requirement), which must still be unique — there is nothing else to discriminate it by.
This also closes #499's dangling `#AssessmentFinding.log` pointer.

Two result semantics get written into doc comments because coverage reports on them: the
SDK's aggregation precedence, which is canonical but undocumented (26a), and
`steps-executed ≤ len(steps)`, present whenever `result` is not `Not Run`, with the failing
step being the last executed one (26b).

### 4.3 Record — `#Evidence` (new `evidence.cue`, stable)

Takes `feat/evidence-digest-semantics-490`'s file as its base, plus the evidence half of
`feat/entity-evidence-reimagine-488`:

- `type` is a descriptor — `#ArtifactType | #URI` — never matched against anything the
  Policy declares (29a). Two documented families: a Gemara artifact type for a citation, a
  URI for any other kind.
- `payload?` or `source?`, keeping the at-least-one rule. `source` is an `#EvidenceMapping`
  carrying a content descriptor: address, `digest?`, `media-type?`, `size?` (29f, #472).
- Digest **recommended** for referenced content, never on inline payloads.
- `originator`, `collector`, `method` from the 488 branch (#491).
- `verified-at`/`verified-by` **dropped** — a verification is an event in the verifier's own
  record, not a field on someone else's citation.
- `id` stays **required** (29b, settled): `#Evidence` is the one type in this design with
  nothing to derive an identity from.
- One type, used at both L5 and L7: audits carry raw evidence such as access reviews, not
  only citations of logs (EV).
- `source.reference-id` resolves under the 1a check (29c); audit evidence pointing into
  logs by ids that don't exist is closed by #499 + D8 (29e).

The move out of `auditlog.cue` is the point: a stable consumer may not depend on a
definition in an experimental file (1h).

### 4.4 Check — `#AuditLog` (auditlog.cue, experimental)

```cue
#AuditLog: {
    #Log
    metadata: type: "AuditLog"
    owner?:   #RACI
    summary:  string

    policy:   #ArtifactMapping                              // was: criteria: [...] (D1)
    coverage: [#RequirementCoverage, ...#RequirementCoverage]
    results:  [#AuditResult, ...#AuditResult]

    "risk-tolerance-met"?: #CheckOutcome                    // 16, 17 — not per-requirement
}

#RequirementCoverage: {
    requirement:    #EntryMapping                        // from effective-requirements
    assessments: [...#EvidenceMapping]                // the assessments read (D8)

    "assessment-present":  #CheckOutcome                 // required
    "plan-bound"?:         #CheckOutcome
    "method-allowed"?:     #CheckOutcome
    "executor-matched"?:   #CheckOutcome
    "parameters-matched"?: #CheckOutcome
    "frequency-met"?:      #CheckOutcome
    "evidence-present"?:   #CheckOutcome
    "evidence-fresh"?:     #CheckOutcome

    "required-methods"?: [...{"method-id": string, outcome: #CheckOutcome}]
}

#CheckOutcome: {
    outcome:  "Pass" | "Fail" | "Not Applicable" | "Undetermined"
    message?: string
}
```

**Where the new types live** ("shared types live in one stable file"):
`#RequirementCoverage` and `#CheckOutcome` are the audit's own, so they stay in
`auditlog.cue`. `#EvidenceExpectation` is the policy's own, so it stays in `policy.cue`; `#ExecutionFacts`
is the evaluation log's own, so it stays in `evaluationlog.cue`.
`#ParameterValue` is referenced by stable `evaluationlog.cue` while its counterpart
`#Parameter` lives in experimental `policy.cue`, so it goes in `primitives.cue` — putting
it anywhere else recreates the 1h violation this change exists partly to fix.

- **`criteria` is renamed to `policy`, not repurposed.** The meaning changed from "artifacts
  I judged against" to "the one policy I audited"; leaving the old name on it would make
  every existing reader silently wrong. Experimental file, minor bump.
- **`AuditResult.criteria-reference` is replaced by a pointer into coverage**, so a result
  states which failed check prompted it — opinion and mechanism agree instead of coexisting.
- **`assessment-present` is required**, which makes coverage complete rather than selective.
- **`risk-tolerance-met` is a named field, not a list.** Risk tolerance `max-severity` (16)
  is audit-level rather than per-requirement, and it gets the same treatment D4 demands of
  coverage: a bare list of checks would make "never checked" and "checked and passed"
  indistinguishable again. This reverses thread 16's "checked as an `AuditResult`"
  precedent and gives 17's chain its end.

This supersedes thread 28's "Complete" verdict. The lineage method traced whether each
field had a beginning, middle and end; it could not see that a field's *type* was too weak
to carry its job.

### 4.5 Plan ↔ log ↔ audit, field by field

No orphans in either direction:

| Plan defines | Log records | Audit checks |
|---|---|---|
| `evaluation-methods[]` | `execution.method-id` | the method used is one the plan allows |
| `...[].required` | `execution.method-id` across the requirement's assessments | every required method was used |
| `...[].executor` | `execution.executor` | matches the plan's when it names one |
| `parameters[]` + `accepted-values` | `execution.parameters[] {id, value}` | every id declared, every value accepted, every plan parameter recorded |
| `frequency-days` | `run.trigger` + `start` | each **required method** ran within the requirement's cadence |
| `evidence-expectation` | `evidence[].collected-at` | the requirement's assessments **collectively** carry ≥1 entry within `valid-for-days` |

`mode` is the one plan field with no record, deliberately — it is implied by `execution.method-id`,
so restating it would be a second copy that can disagree.

Constraints appear nowhere in this table on purpose. A constraint is folded into the
requirement set by resolution, so an assessment run under a plan is already run against the
constrained requirement; checking the requirement *is* checking the constraint. Tracking
them separately would be the "Layer 5/6 tracking" claim thread 24 found to be false.

## 5. Where each rule lives

**CUE (intra-document).** 1a membership with 7b self-reference; `requirement-id` unique
across a policy's `assessment-plans`; `assessment-present` required;
coverage coherence (`evidence-fresh` implies `evidence-present`; plan-dependent checks
absent or `Not Applicable` when `plan-bound` failed); D8's 4-tuple uniqueness; `policy` →
`effective-requirements`; `plan` present ⇔ `execution` present;
evidence at-least-one plus the #490 descriptor coupling; `steps-executed` bounds (26b).

**Tool spec (cross-document or procedural).** **Coverage completeness** — exactly one entry
per effective requirement, none extra; the rule this design exists for and the one CUE can
never hold. Check conditionality: the plan defines it → coverage reports it. Version pins
equal the fetched `metadata.version` (1b). Cadence; parameter values within
`accepted-values`; cadence, per required method against the requirement's
`frequency-days`; evidence freshness, evaluated across the requirement's assessments
collectively; the digest spot check, with undigested referenced evidence reported as not
spot-checkable; required-method usage; executor match.

**Doc comments.** The converter publishes the **whole** comment as the field's API
description, not the first paragraph — an earlier version of this document said otherwise
and was wrong, which is how `_|_` once reached the golden file. A note that should not be
published goes in the validation section at the bottom of the definition, detached by a
blank line.

## 6. Fixtures and tests

**Repairs required by this change's rules:** declare the missing mapping-references in
`good-ccc.json`, `nested-good-ccc.yaml`, `good-evaluation-log-unstarted.yaml` and
`pvtr-baseline-scan.yaml` (1a); `metadata.version` on the two CCC fixtures (1b);
`good-evaluation-log-unstarted.yaml:21` → `Unknown`, matching the canonical precedence
(26a); `good-audit-log.yaml`'s entry-id (29e).

**The scenario.** Nothing in the repo exercises a plan at all (22a), so one coherent
fixture set threads the four types by real ids: a Policy with real assessment plans
(methods with `required` and `executor`, parameters with `accepted-values`,
`frequency-days`, an `evidence-expectation`); an EvaluationLog that cites that policy,
carries `effective-requirements`, and records an `execution` per assessment, with evidence; an AuditLog auditing that policy with complete
coverage. The existing enforcement log chains onto the same evaluation log, closing #499's
pointer with real data.

**Negative fixtures**, one per new CUE rule: dangling reference; `execution` without a
plan; `plan` without `execution`; two plans for one `requirement-id`; coverage missing
`assessment-present`; a duplicate identity 4-tuple;
plus the seven digest and media-type cases from #490. Each registered in
`test/schema_test.go`.

**Gates on every commit:** `make test`, `cuefmtcheck`, `lintcue`, `genopenapi`,
`breaking-check` against the v1 baseline with `.oasdiff-allow` entries for the intentional
experimental-file replacements. `breaking-check` matters more than usual — a CUE field
alias silently drops a field from the generated OpenAPI (`metadata.cue:51` uses one), and
it is the only gate that catches it.

## 7. Versioning impact

| File | Status | Change class |
|---|---|---|
| `policy.cue` | experimental | Minor — `#AssessmentPlan` restructuring allowed |
| `auditlog.cue` | experimental | Minor — `criteria` → `policy` allowed |
| `evaluationlog.cue` | **stable** | Additive fields + conditional tightenings |
| `primitives.cue` (was `mapping_inline.cue`) | **stable** | Additive — `#EvidenceMapping` gains descriptor fields |
| `metadata.cue` | **stable** | Tightening — an optional field becomes required (1b) |
| `evidence.cue`, `primitives.cue` | new | `#Evidence` leaves experimental `auditlog.cue` (1h) |

No change requires a v2. Tightenings on stable files are permitted in v1 by project
definition; making an existing optional field required is a tightening, not a new required
field.

## 8. Sequencing

Each commit passes the full gate set.

1. **1a + 7b** — reference membership at every artifact root, with fixture repairs.
   Prerequisite; may land on its own if needed.
2. **1b** — `Metadata.version` required, with repairs.
3. Merge **`worktree-fix-entrymapping-ids`** (#499).
4. **Evidence** — `primitives.cue`, `evidence.cue`, the `#EvidenceMapping` rework, the 488
   evidence half, `verified-at`/`verified-by` dropped (§4.3).
5. **Define** — `#AssessmentPlan` and `#EvidenceExpectation` (§4.1).
6. **Record** — `#AssessmentLog` additions, `#EffectiveRequirement`, the snapshot, D8's
   uniqueness (§4.2).
7. **Check** — `#AuditLog`, `#RequirementCoverage`, `#CheckOutcome`, `risk-tolerance-met`
   (§4.4).
8. **Scenario fixtures** (§6).
9. **Tool spec text** for the cross-document rules in §5.

## 9. Notes carried into implementation

No open items remain. Two things worth carrying forward:

- **Thread 28's verdict is superseded**, and the reason is worth recording: the lineage
  method could not see a type that was too weak for its job, only a field with no consumer.
- **One deliberate deviation from the audit's stated fix.** Thread 21 proposes `method` as
  an `#EntryMapping` to the `#AcceptedMethod`; this spec uses a bare `method-id: string`.
  The plan is already identified by `plan`, so the method is scoped by it and a full
  reference would restate what is known.
