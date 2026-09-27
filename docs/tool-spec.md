# Gemara tool specification

**Status: draft.** This document is destined for
[github.com/gemaraproj/website](https://github.com/gemaraproj/website), which holds
Gemara's prose; it lives here while it is drafted. The reference pages the website
generates from the CUE schema are not a substitute for it: they describe shapes,
and this describes obligations.

## Who this is for

A tool that produces or consumes Gemara documents — a scanner emitting evaluation
logs, a gate emitting enforcement logs, an auditor's tooling emitting audit logs, a
dashboard reading any of them. Validating a document against the schema is necessary
and not sufficient. This document states the rules a conforming tool must implement
itself.

## The division of labour

CUE validates **one document at a time**, structurally, without retrieving anything.
Everything else is a tool's job. Three kinds of rule fall outside the schema, and
each falls outside for its own reason:

| kind | why the schema cannot hold it | example |
|---|---|---|
| cross-document | validation sees one file; the other document is elsewhere, at a pinned version | the `method-id` an assessment ran must exist in the plan it cites |
| procedural | requires retrieval, hashing, clock arithmetic — none of which CUE does | verifying a `digest` against what is at a `download-url` |
| judgement | there is a defensible answer either way, and a person is accountable for it | whether unmet requirements are material enough to fail an audit |

Where a rule is stated in a `.cue` doc comment as documented-but-not-enforced, this
document says what implementing it means. Where a rule is enforced, section 1 says so
plainly, so tools do not re-implement it.

## 1. What the schema already guarantees

A document that validates has these properties. A tool may rely on them without
checking, and should not report them as its own findings.

**Identity and uniqueness.** Within one document: mapping-reference ids, group ids,
control ids, requirement ids, guideline ids, principle, capability, threat, risk and
vector ids, lexicon term ids, risk ranks, mapping ids. Within one policy: assessment
plan ids, evaluation and enforcement method ids (across every site that declares
one), parameter ids within their plan, and **one plan per requirement**. Within one
evaluation log: a control is evaluated once, and within one control evaluation an
assessment is unique on the tuple (requirement, plan-id, method-id), where a missing
plan or method counts as its own value rather than colliding. Within one enforcement
log: action ids, and finding ids across every action's justification.

**Reference integrity inside the document.** Every `reference-id` names a
`mapping-reference` the document declares, or the document's own `metadata.id`. This
is enforced for logs, for a catalog's `extends` and `imports`, for a policy's imports
and its plans' requirements, and for an audit's policy, verified requirements,
evidence sources, and findings' requirement and risk pointers. It does **not** mean
the referenced artifact exists or contains that entry — see section 2.

**Evidence integrity claims.** Every `#Evidence` carries an `id`, a `type` and
`collected-at`, and carries a payload, a source, or both. An inline payload and a
`source.download-url` are mutually exclusive: carried content has no address, and
addressed content is referenced rather than carried. A `download-url` requires a
`digest`. A digest matches the `algorithm:encoded` grammar, and registered algorithms
are checked for their own encoding and length.

**Evaluation.** An assessment whose `result` is anything other than `Not Run`,
`Unknown` or `Not Applicable` records a `start`, since it ran; `steps-executed` never
exceeds the number of steps; a cited plan obliges `#ExecutionFacts`, and without a
plan there are none to record.

**Enforcement.** A `Clear` action carries no justification, because it found nothing
to act on. Every other disposition carries a justification that justifies something:
findings, or the exceptions that stood the action down. Each finding in a
justification carries an `id`, the evaluation log entry it came from, and a
`severity`.

**Audit.** The audit cites exactly one policy and states an `opinion`. Every
verification entry states an `outcome`, the requirement it concerns, the
`effective` requirement as resolved, and `evidence-present`. An entry that claims
`evidence-present: Satisfied` names evidence or assessments. `effective.plan-id` and
the `plan` verification oblige each other, and a plan verification enumerates at
least one method, each of which answers `allowed`, `used` and `cadence-met`. A
finding names a requirement the audit verified, or a risk. A finding determined
`Not Compliant` states a `severity`.

## 2. Cross-document rules a tool must check

Each rule names the documents it needs. A tool that cannot retrieve the second
document reports the check as **unperformed**, not as passed.

### 2.1 Evaluation log against the policy it cites

1. **The plan exists.** `#AssessmentLog.plan` resolves to an `#AssessmentPlan` in the
   cited policy at the cited version, and that plan governs the requirement the
   assessment names.
2. **The method is one the plan accepts.** `#ExecutionFacts.method-id` appears in
   that plan's `evaluation-methods`. Method ids are unique across a policy, so no
   further qualifier is needed.
3. **The executor matches.** Where the plan's accepted method names an `executor`,
   compare it against `#ExecutionFacts.executor` by authoritative identifier — a
   package path, an image reference, an account — never by display name or internal
   short name.
4. **Parameter values are permitted.** Each `#ParameterValue` names a parameter the
   plan declares, and its value falls within that parameter's `accepted-values`
   where the parameter constrains them.
5. **Cadence.** Compare the assessment's `start` against the plan's
   `frequency-days`. Cadence belongs to the requirement, not to a plan instance or a
   method, so the window is the same for every method the plan requires — but each
   required method must have run inside it.

### 2.2 Audit against the policy and the catalogs beneath it

6. **Completeness.** The policy determines the set of requirements in scope. There
   must be exactly one verification entry per requirement that set contains. The
   schema cannot check this, because the set lives in the policy and the catalogs it
   imports; it is the single most important check a tool performs on an audit, since
   an audit's value rests on being systematic rather than selective.
7. **The effective requirement is faithful.** `effective.text`, `applicability` and
   `constraints` must equal what the cited policy, applied to the catalog at the
   version the policy pins, actually produces. A policy selects, constrains and
   parametrizes; it never rewrites catalog content, so a text that cannot be derived
   that way is a defect in the audit, not a difference of opinion.
8. **The plan is the policy's.** `effective.plan-id` names a plan in the cited policy
   that governs that requirement.

### 2.3 Audit against the evaluation logs it read

9. **Assessment pointers resolve.** Each `assessments` entry resolves to an entry
   that exists in the log it names, and that entry concerns the same requirement as
   the verification entry citing it.
10. **Freshness.** `evidence-fresh` must be consistent with evidence `collected-at`
    against the plan's `evidence-requirements.valid-for-days` where a plan sets one.
    Where none does, the auditor's basis for the judgement should be stated in a
    finding; a tool cannot compute it and must not assume a default.

An audit may be built on evidence that is not a Gemara document at all. That is a
supported case, not a degradation: the entry answers `evidence-present`,
`evidence-fresh` and `outcome`, and nothing more. What a tool must not accept is the
*silent* version of it — where a policy binds a plan and the audit omits plan
conformance. The schema forces the `plan` block to exist in that case; a tool should
check that its determinations reflect what the audit actually had.

### 2.4 Enforcement log against the evaluation log it reacted to

11. **Findings resolve.** Each justification finding's `log` entry exists in the log
    it names and concerns the requirement the finding claims. Note that the
    enforcement log itself is not expected to validate this: enforcement records what
    a gate did, and checking policy semantics is the audit's role. A tool that reads
    both documents can and should check it anyway.
12. **Exceptions resolve.** Gemara does not model an exception; a `Tolerated` action
    names external ones through its mapping-references. A tool that has access to the
    exception register should confirm the exception exists, covers this subject, and
    was in force when the action ran.

### 2.5 Integrity of anything cited

13. **Digest verification.** For a citation with a `download-url` and a `digest`:
    retrieve the octet stream, hash it exactly as delivered, and compare. Report
    exactly one of three outcomes per citation — **verified**, **integrity-failure**,
    or **unverifiable** (retrieval failed, or the algorithm is unregistered or
    unimplemented). `sha256` is the MUST-support floor for a conforming verifier;
    `sha512` and `blake3` are MAY. An unrecognized algorithm that complies with the
    grammar is unverifiable rather than invalid.
14. **Absence is not assurance.** A citation with no digest makes no integrity claim.
    It is never reported as verified.

### 2.6 Policy implementation timelines

15. **Interpretation, not validation.** `#ImplementationPlan`'s
    `evaluation-timeline` and `enforcement-timeline` are communicated rather than
    enforced, and nothing references them. They carry interpretation a later reader
    cannot reconstruct: that enforcement was not yet due rather than skipped, and
    whether an audit's conformance statement falls inside the window the policy was
    actually in effect. A tool that has the policy and a log should surface that
    comparison; it should not treat a log outside the window as invalid.

## 3. Aggregation

Every roll-up in Gemara is documented rather than enforced, because a producer may
have information the schema cannot see. A tool that aggregates must follow these, and
a tool that reads an aggregate should check it.

**`#Result`** — assessment, control evaluation, log:
`Failed > Unknown > Needs Review > Passed > Not Applicable`, and `Not Run` never
overwrites another value.

**`#Disposition`** — action to log:
`Enforced > Tolerated > Undetermined > Clear`. Anything enforced makes the log
`Enforced`; `Clear` survives only when every action was. The strongest action
dominates, because a disposition records what the enforcement layer *did*.

**`#VerificationLog.outcome`** — the requirement's determination. A check that came
out `Not Satisfied` is expected to show in the outcome, since a requirement that is
unevidenced, stale or unverified against its plan cannot be determined `Compliant`.
`Not Applicable` survives only where every check was.

**`#Opinion`** — the audit as a whole. This is **not** a roll-up. Whether unmet
requirements are material enough to sink the audit is the auditor's judgement, which
no precedence over per-requirement outcomes can compute: three `Not Compliant`
requirements out of two hundred may be `Passed with Conditions`, while one on the
requirement the target exists to meet is `Failed`. A tool must not derive it, and
must not overwrite it. Anything other than `Passed` is expected to be explained by
findings.

**Lifecycle correspondence.** A `Tolerated` action's findings are `Waived`. No other
correspondence between a disposition and a finding's lifecycle holds: a blocked merge
is `Enforced` with its finding still `Open`, because the gate prevented harm without
fixing anything, while a remediation is also `Enforced` and does leave its finding
`Resolved`.

## 4. Rules the schema leaves to tools on purpose

These could be expressed in CUE and are deliberately not, for the reason given.

16. **A justification names at least one finding.** Requiring a non-empty list in
    CUE forces an element into existence, which demands concrete values for that
    element's required fields and produces an error about a finding nobody wrote. A
    tool should reject a justification whose `findings` and `exceptions` are both
    empty lists.
17. **A questionable outcome is explained.** Nothing in a verification entry reasons.
    An `outcome` of `Not Compliant` or `Undetermined` should be matched by a finding
    naming that requirement. The schema requires findings to name a verified
    requirement, but not the converse, since a positive finding is legitimate and an
    auditor may consider an outcome self-explanatory.
18. **Applicability agrees.** The same applicability appears in the catalog
    requirement, in the audit's `effective`, and on the assessment log. The copies
    are deliberate — an audit is a point-in-time attestation, and the log's copy aids
    execution — but a tool holding all three should report disagreement.

## 5. Producer guidance

**Context flows downstream, never upstream.** An evaluation log must be readable on
its own, because most are never audited: it names its plan as a mapping, policy and
plan together. An audit may compress the same pointer to a bare `effective.plan-id`
only because it pins `policy` once at the document level. Compression is safe when
the context is upstream and pinned, and a dangling pointer when it is downstream and
optional.

**Mint an id only where nothing derives identity.** A verification entry is
identified by its requirement, so it has no id. An audit finding has one, because
several findings may concern one requirement and nothing else distinguishes them. An
`#Evidence` has one, because a reference to it may be wanted later and nothing else
would serve.

**Record what the run chose, not what the method is.** `#ExecutionFacts` names the
method, and the executor and parameter values the run used, because each could differ
from what the plan said. It does not copy the method's `type` or `mode`: `method-id`
resolves those, and a second copy could only disagree.

**Record an actor only when it differs from the log's author.**
`#ExecutionFacts.executor` and `#Evidence.collector` exist for the third-party case.
Where the log's `metadata.author` did the work, saying so again is a second copy of
one fact.

## 6. Standards alignment

Gemara borrows vocabulary deliberately, so that a projection is a rename rather than
a reinterpretation. Cited by term rather than by clause number: the numbering must be
verified against the standards themselves, which are paywalled, so everything here is
paraphrase.

| Gemara | standard | the borrowed idea |
|---|---|---|
| `#AuditLog.policy` | ISO 19011, *audit criteria* | criteria are a stated set of requirements, which is why an audit cites one versioned policy rather than a list of artifacts |
| `#Finding` | ISO 19011, *audit finding* | the result of evaluating collected evidence against the criteria; indicates conformity as readily as nonconformity |
| `#Evidence` | ISO 19011, *audit evidence* | records and statements of fact that are relevant and **verifiable** — the reason a citation carries a digest |
| `#ComplianceStatus` | ISO 19011, conformity/nonconformity | compliance is stated against criteria, one requirement at a time |
| `#Opinion` | ISO 19011, *audit conclusion* | reached after weighing all findings against what the audit set out to establish; hence a judgement |
| `#Opinion`'s values | audit opinion practice | unqualified → `Passed`, qualified → `Passed with Conditions`, adverse → `Failed`, disclaimer of opinion → `Undetermined` |
| `severity` on enforcement findings | ISO/IEC 17021-1, grading of nonconformities | major/minor grading is what `#Severity` carries |
| `#Finding`, `#ComplianceFinding` | OCSF Compliance Finding | `lifecycle` → `status_id`, `description`/`severity` → `finding_info.desc`/`severity_id`, `recommendations` → `remediation`, `status` → `compliance.status` |
| `#Evidence.type` | in-toto `predicateType` | an evidence type names the shape of what was collected |
| `#Policy` | OSCAL profile | a policy selects, constrains and parametrizes catalog content; it never rewrites it |

## 7. Open items

- **A digest is one value, not a set.** The rationale is recorded on `#Digest`. The
  trigger that would reopen it is handing a subject or materials list to an in-toto
  verifier, whose `ResourceDescriptor` mandates a `DigestSet`. The place for it then
  is a content descriptor primitive, not a wider `#Digest`.
- **A content descriptor primitive.** `#EvidenceMapping` carries `download-url`,
  `digest`, `size` and `media-type` as local fields. They are one reusable idea and
  should become one.
- **Clause numbers** in section 6 need verifying against the standards.
