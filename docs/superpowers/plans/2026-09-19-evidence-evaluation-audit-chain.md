# Evidence, Evaluation and Audit Chain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Gemara's define → record → check chain so an audit can state, against a named Policy, that every requirement was covered and every plan was followed.

**Architecture:** Four CUE types change together. `#AssessmentPlan` (Policy, L3) gains checkable commitments; `#AssessmentLog` (EvaluationLog, L5) records what a run actually used and the effective requirement set it ran against; `#Evidence` moves to its own stable file with content descriptors; `#AuditLog` (L7) takes a single versioned Policy as criteria and gains per-requirement coverage with dedicated named check fields. Rules that a single document can check live in CUE; cross-document rules go to the tool spec.

**Tech Stack:** CUE (schema), Go (test harness in `test/`, converter in `cmd/`), Make targets for gates, oasdiff via `make breaking-check`.

**Spec:** `docs/superpowers/specs/2026-09-19-evidence-evaluation-audit-chain-design.md`

## Global Constraints

- **Branch:** `fix/cue2openapi`. Do not create a new branch. Do not push or open a PR unless asked.
- **Every commit passes all gates**, in this order:
  `make test` · `make cuefmtcheck` · `cue eval . --all-errors` · `make genopenapi` · `make breaking-check`
- **`make test`, not just the fixture suite.** `make test` runs *two* Go suites: `test/` (fixture validation) and `cmd/` (the CUE→OpenAPI converter, including `TestGoldenDocument`). Running only `cd test && go test` leaves converter regressions undetected — it did exactly that in Task 1.
- **`make lintcue` is broken in this environment** and it is not this change's fault: the Makefile passes `--verbose`, which `cue` v0.17.1 rejects. Verified failing on unmodified `6e463c0`. Run `cue eval . --all-errors` instead, which is the real check. Do **not** fix the Makefile inside this plan — it is unrelated scope and would muddy the diff.
- **Doc-comment changes make `cmd/internal/cmd/testdata/openapi.golden.yaml` stale.** The converter publishes the first paragraph of every doc comment, so almost every task in this plan will change it. Regenerate with `cd cmd && UPDATE_GOLDEN=1 go test -count=1 ./internal/cmd/projection -run TestGoldenDocument`, then **read the resulting diff** and confirm it contains only the descriptions and fields your task intended to change. Never regenerate a golden file without inspecting what moved — that is the only thing keeping it honest.
- **`-count=1` is mandatory** on Go tests: the tests read `*.cue` at runtime and Go's cache cannot track that, so cached passes go stale against edited schemas.
- **Stability tags are law.** `evaluationlog.cue`, `mapping_primitives.cue`, `metadata.cue`, `collections.cue`, `entities.cue` are `@gemara(status="stable")` — additive fields and tightenings only, never a rename or removal. `policy.cue`, `auditlog.cue`, `enforcementlog.cue` are `@gemara(status="experimental")` — any change is a minor.
- **Doc comments state the rule in the first sentence, and contain nothing internal.** The OpenAPI converter publishes a field's *entire* attached doc comment as its API description — every paragraph, not just the first (an earlier version of this plan claimed otherwise; that was wrong). An implementation note belongs in a separate comment block detached from the field by a blank line, or it becomes public API text.
- **Attribute syntax:** this repo uses `@gemara(status="...")`. The `feat/evidence-digest-semantics-490` branch predates that rename and uses `@status("...")` — convert when copying files from it.
- **No new minted ids.** Identity is derived from existing fields. The one exception is `#Evidence.id`, which stays required because Evidence has nothing to derive from.
- **CUE version of "unique":** a struct comprehension keyed by the value, e.g. `_uniquePlanIds: {for i, p in plans {(p.id): i}}`. Duplicate keys with different values conflict and fail validation.
- **Commit style:** Conventional Commits, with these trailers, in this order:
  ```
  Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
  Assisted-by: Claude (model: claude-opus-5)
  ```

---

### Task 1: Reference membership (threads 1a, 7b)

Every `reference-id` in a document must name one of the document's declared `metadata.mapping-references`, or the document's own `metadata.id` (self-reference). Today only the uniqueness of mapping-reference ids is checked (metadata.cue:62-63), so a pointer to nothing validates. Every pointer added later in this plan depends on this rule.

**Files:**
- Modify: `auditlog.cue`, `evaluationlog.cue`, `policy.cue`, `controlcatalog.cue`
- Modify: `mapping_inline.cue:9-12` (doc comment on `#MappingReference.id`)
- Modify: `test/test-data/good-ccc.json`, `test/test-data/nested-good-ccc.yaml`, `test/test-data/good-evaluation-log-unstarted.yaml`, `test/test-data/pvtr-baseline-scan.yaml`
- Create: `test/test-data/bad-evaluation-log-undeclared-reference.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: the `_refValidation` pattern every later task reuses, and the guarantee that `AuditLog.policy`, `Evidence.source.reference-id` and coverage pointers resolve.

- [ ] **Step 1: Write the failing fixture**

Create `test/test-data/bad-evaluation-log-undeclared-reference.yaml` by copying `good-evaluation-log-unstarted.yaml` and changing one `reference-id` to a name that is not declared:

```yaml
# copy of good-evaluation-log-unstarted.yaml with one change:
# evaluations[0].control.reference-id: OSPS-B  ->  NOT-DECLARED-ANYWHERE
```

Add the row to the `tests` table in `test/schema_test.go`, in the `// EvaluationLog — negative` block:

```go
{"evaluation log citing an undeclared reference", "./test-data/bad-evaluation-log-undeclared-reference.yaml", "#EvaluationLog", true, ""},
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/evaluation_log_citing_an_undeclared_reference' ./...`
Expected: FAIL with "expected validation error, got nil" — nothing checks membership yet.

- [ ] **Step 3: Repair the four fixtures that are wrong today**

These pass now only because the rule does not exist. `good-ccc.json` and `nested-good-ccc.yaml` declare no `mapping-references` at all yet cite CCM, CSF, ISO-27001, NIST-800-53 and CCC; `good-evaluation-log-unstarted.yaml` and `pvtr-baseline-scan.yaml` cite `OSPS-B` without declaring it. Add the missing declarations to each file's `metadata`:

```yaml
metadata:
  # ...existing...
  mapping-references:
    - id: OSPS-B
      title: Open Source Project Security Baseline
      version: "2025.02"
      url: https://baseline.openssf.org/versions/2025-02-25.html
```

Use the ids each file actually cites. Every cited id must appear.

- [ ] **Step 4: Add the rule to each artifact root**

The declared set is the document's mapping-reference ids plus its own `metadata.id`. In `evaluationlog.cue`, inside `#EvaluationLog`:

```cue
import "list"

#EvaluationLog: {
	// ...existing fields...

	// Every reference-id in this log MUST name one of its declared
	// mapping-references, or the log's own metadata.id (self-reference).
	let _declaredRefIds = [metadata.id] + [
		if metadata["mapping-references"] != _|_
		for r in metadata["mapping-references"] {r.id},
	]
	for i, e in evaluations {
		_refValidation: "evaluations-\(i)": _declaredRefIds & list.Contains(e.control."reference-id")
		for j, a in e."assessment-logs" {
			_refValidation: "evaluations-\(i)-\(j)": _declaredRefIds & list.Contains(a.requirement."reference-id")
		}
	}
}
```

Apply the same shape to `#AuditLog` (over `criteria[].reference-id`, `results[]."criteria-reference"."reference-id"`, and `results[].evidence[].source."reference-id"`), to `#Policy` (over `imports` and `adherence."assessment-plans"[]."requirement-id"` once it is an `#EntryMapping` — at this point it is still a bare string, so skip that site and pick it up in Task 5), and to `#ControlCatalog` (over the sites the CCC fixtures cite).

If CUE rejects the chained `if ... for ...` comprehension clause, fall back to a guarded `if metadata["mapping-references"] != _|_ { ... }` block **plus** a rule that `mapping-references` must be present whenever any reference-id is set — do not leave the check behind a guard that silently disables it, which is the 14a defect.

- [ ] **Step 5: Add the self-reference doc comment (7b)**

In `mapping_primitives.cue`, `#MappingReference.id`'s comment already mentions self-reference; make the rule its first sentence:

```cue
// id identifies this mapping reference within the artifact; every reference-id
// in the artifact MUST equal one of these ids or the artifact's own metadata.id.
```

- [ ] **Step 6: Run the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`
Expected: all PASS, including the new negative fixture and the four repaired fixtures.

- [ ] **Step 7: Commit**

```bash
git add auditlog.cue evaluationlog.cue policy.cue controlcatalog.cue mapping_primitives.cue test/
git commit -m "$(cat <<'EOF'
fix(schema): require every reference-id to name a declared mapping-reference

Only the uniqueness of mapping-reference ids was checked, so a document
could cite a reference it never declared and still validate. Add the
membership check at each artifact root, allowing the document's own
metadata.id as a self-reference, and repair the four fixtures that were
relying on the gap.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 2: Required artifact version (thread 1b)

`#MappingReference.version` is required (mapping_inline.cue:18), so every citation pins a version — but the cited artifact need not declare one (`Metadata.version` is optional, metadata.cue:39), so a pin can point at nothing. Task 9 makes an audit cite a *versioned* Policy, which is only meaningful once this holds.

> Note from user: This is an unacceptable changes. We would need to only make the version required in certain cases. Like a document of a certain status. MappingDocuments only want to crosswalk publish artifacts anyway.

**Files:**
- Modify: `metadata.cue:38-39`
- Modify: `test/test-data/good-ccc.json`, `test/test-data/nested-good-ccc.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: Task 1's repaired fixtures.
- Produces: `metadata.version` guaranteed present on every artifact.

- [ ] **Step 1: Write the failing fixture**

Create `test/test-data/bad-control-catalog-missing-version.yaml` — a minimal valid ControlCatalog with `metadata.version` omitted. Copy `good-ccc.yaml` and delete its `version:` line.

Add to `test/schema_test.go` in the `// ControlCatalog — negative` block:

```go
{"control catalog without a version", "./test-data/bad-control-catalog-missing-version.yaml", "#ControlCatalog", true, ""},
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/control_catalog_without_a_version' ./...`
Expected: FAIL — `version` is optional, so the document validates.

- [ ] **Step 3: Make the field required**

In `metadata.cue`:

```cue
	// version is the version identifier of this artifact; it MUST be present so
	// that a MappingReference pinning this artifact has a value to match.
	version: string
```

- [ ] **Step 4: Repair the two fixtures without a version**

Add `version: "2025.1"` (or the value each file's content implies) to `good-ccc.json` and `nested-good-ccc.yaml`.

- [ ] **Step 5: Run the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

`breaking-check` will flag `version` becoming required. This is a **tightening**, permitted in v1 by project definition — an existing optional field becoming required is not "adding a required field". Add an entry to `.oasdiff-allow` naming this change, with a one-line reason.

- [ ] **Step 6: Commit**

```bash
git add metadata.cue .oasdiff-allow test/
git commit -m "$(cat <<'EOF'
fix(schema): require metadata.version on every artifact

MappingReference.version is required, so every citation pins a version,
but the cited artifact could omit one entirely — a pin with nothing to
match. Make version required and give the two fixtures that lacked one
a value.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 3: Merge derived identity for log entries (#499)

`worktree-fix-entrymapping-ids` gives `#EntryMapping` targets real identity with no new stored ids. Task 7 widens one of its rules, so it must land first.

**Files:**
- Modify: merge commit only.

**Interfaces:**
- Produces: `_uniqueEvaluatedControls` and `_uniqueAssessedRequirements` in `evaluationlog.cue`; `_uniquePlanIds`, `_uniqueMethodIds` (policy-wide, across all three method sites) and `_uniqueParameterIds` in `policy.cue`; `id` on `#ActionResult` and `#AssessmentFinding`.
- **`_uniqueMethodIds` being policy-wide is what makes a bare `method-id` string unambiguous in Task 6** — it needs no catalog or plan qualifier.

- [ ] **Step 1: Merge the branch**

```bash
git merge --no-ff worktree-fix-entrymapping-ids
```

- [ ] **Step 2: Resolve conflicts against Task 1's work**

Both branches touch `evaluationlog.cue` and `policy.cue`. Keep both sets of rules: Task 1's `_refValidation` blocks and the branch's `_unique*` blocks are independent and must coexist.

- [ ] **Step 3: Run the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`
Expected: all PASS, including the branch's new negative fixtures (`bad-evaluation-log-duplicate-requirement.yaml`, `bad-policy-duplicate-plan-id.yaml`, and the rest).

- [ ] **Step 4: Commit**

The merge commit is the commit. If conflicts were resolved, amend with a message naming what was kept:

```bash
git commit --amend -m "$(cat <<'EOF'
merge: derived identity for log entries (#499)

Brings composite-key identity for ControlEvaluation, AssessmentLog,
plans, methods and parameters. Kept alongside the reference-membership
rules added on this branch; the two are independent.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 4: Evidence moves to its own stable file (EV, 29a–29f, 1h)

`#Evidence` sits in experimental `auditlog.cue` while stable `evaluationlog.cue` consumes it — a stable type depending on an experimental one. It moves to `evidence.cue`, tagged stable, with the content-descriptor work from `feat/evidence-digest-semantics-490` and the evidence half of `feat/entity-evidence-reimagine-488`.

**Files:**
- Create: `primitives.cue`, `evidence.cue`
- Modify: `auditlog.cue` (remove `#Evidence`, `#_EvidenceStrict`, `#EvidenceType`), `mapping_primitives.cue` (`#EvidenceMapping` gains descriptor fields)
- Create: seven negative fixtures from the #490 branch
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: Task 1's membership rule, which is what makes `Evidence.source.reference-id` resolve (29c).
- Produces: `#Evidence`, `#EvidenceType`, `#EvidenceMapping`, `#URI`, `#Digest`, `#ContentDescriptor`. `#ParameterValue` is added to `primitives.cue` in Task 6.

- [ ] **Step 1: Take the files from the #490 branch**

```bash
git checkout feat/evidence-digest-semantics-490 -- primitives.cue evidence.cue
git checkout feat/evidence-digest-semantics-490 -- \
  test/test-data/bad-audit-log-digest-sha256-length.yaml \
  test/test-data/bad-audit-log-digest-sha256-uppercase-hex.yaml \
  test/test-data/bad-audit-log-digest-sha512-length.yaml \
  test/test-data/bad-audit-log-download-url-invalid.yaml \
  test/test-data/bad-audit-log-media-type-malformed.yaml \
  test/test-data/bad-audit-log-media-type-without-download-url.yaml \
  test/test-data/bad-audit-log-size-without-download-url.yaml \
  test/test-data/good-audit-log-digest-profile.yaml
```

- [ ] **Step 2: Convert the attribute syntax**

Both copied `.cue` files open with `@status("stable")`. Change each to:

```cue
@gemara(status="stable")
```

- [ ] **Step 3: Apply this design's decisions to `evidence.cue`**

Four edits to the copied file:

1. `type` becomes a descriptor with a closed shape — replace `#EvidenceType: #ArtifactType | string` with:

```cue
// EvidenceType describes what this evidence is; it is never matched against
// anything a Policy declares. Two families: a Gemara artifact type (an audit
// citing an EvaluationLog or EnforcementLog, verifiable against that schema),
// or a URI naming any other kind.
#EvidenceType: #ArtifactType | #URI @go(-)
```

2. Keep `id: string` **required**. Do not relax it — Evidence is the one type here with nothing to derive an identity from.
3. Add the 488 branch's collection fields:

```cue
	// originator is the party that produced the evidence content.
	originator?: #Actor

	// collector is the party that gathered it into this log.
	collector?: #Actor

	// method describes how the evidence was collected.
	method?: #CollectionMethod
```

Copy `#CollectionMethod` from `git show feat/entity-evidence-reimagine-488:entities.cue`.

4. Do **not** add `verified-at` or `verified-by` from the 488 branch. A verification is an event in the verifier's own record, not a field on someone else's citation.

- [ ] **Step 4: Remove the old definitions from `auditlog.cue`**

Delete `#Evidence`, `#_EvidenceStrict` and `#EvidenceType` (auditlog.cue:79-115). `#AuditResult.evidence` and `#AssessmentLog.evidence` keep their existing declarations and now resolve to `evidence.cue`.

- [ ] **Step 5: Register the new fixtures**

Add to `test/schema_test.go` in the `// AuditLog — negative` block (one row per bad fixture), and one positive row:

```go
{"audit log with a full digest profile", "./test-data/good-audit-log-digest-profile.yaml", "#AuditLog", false, ""},
{"audit log digest with wrong sha256 length", "./test-data/bad-audit-log-digest-sha256-length.yaml", "#AuditLog", true, ""},
{"audit log digest with uppercase hex", "./test-data/bad-audit-log-digest-sha256-uppercase-hex.yaml", "#AuditLog", true, ""},
{"audit log digest with wrong sha512 length", "./test-data/bad-audit-log-digest-sha512-length.yaml", "#AuditLog", true, ""},
{"audit log evidence with invalid download url", "./test-data/bad-audit-log-download-url-invalid.yaml", "#AuditLog", true, ""},
{"audit log evidence with malformed media type", "./test-data/bad-audit-log-media-type-malformed.yaml", "#AuditLog", true, ""},
{"audit log evidence media type without download url", "./test-data/bad-audit-log-media-type-without-download-url.yaml", "#AuditLog", true, ""},
{"audit log evidence size without download url", "./test-data/bad-audit-log-size-without-download-url.yaml", "#AuditLog", true, ""},
```

- [ ] **Step 6: Run the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

Watch `make genopenapi` output: moving a definition between files must not change the generated schema for `#Evidence`. If `breaking-check` reports a *removed* field, the move dropped something — fix rather than allow-list it.

- [ ] **Step 7: Commit**

```bash
git add primitives.cue evidence.cue auditlog.cue mapping_primitives.cue entities.cue test/
git commit -m "$(cat <<'EOF'
feat(evidence): give Evidence its own stable file with content descriptors

Stable evaluationlog.cue consumed #Evidence from experimental
auditlog.cue. Move it to evidence.cue with #URI, #Digest and
#ContentDescriptor in primitives.cue, add originator/collector/method,
and constrain EvidenceType to an artifact type or a URI so a typo
cannot pass as a kind. verified-at/verified-by are deliberately not
carried over: a verification belongs in the verifier's own record.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 5: Define — `#AssessmentPlan` (threads 22b, 23, 29d, 21, 24)

The plan's four unverifiable commitments become checkable ones.

**Files:**
- Modify: `policy.cue:107-130` (`#AssessmentPlan`, `#AcceptedMethod`), `policy.cue:156-164` (`#Constraint`)
- Create: `test/test-data/bad-policy-duplicate-requirement-plan.yaml`
- Modify: `test/test-data/good-policy.yaml`, `test/test-data/good-security-policy.yml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: `#EntryMapping` (mapping_inline.cue:48).
- Produces: `#AssessmentPlan` with `"requirement-id": #EntryMapping`, `"frequency-days": int`, `"evidence-expectation"?: #EvidenceExpectation`; `#EvidenceExpectation: {description: string, "valid-for-days": int}`. Task 9's coverage joins on `requirement-id`.

- [ ] **Step 1: Write the failing fixture**

Create `test/test-data/bad-policy-duplicate-requirement-plan.yaml`: a Policy with two entries in `adherence.assessment-plans` whose `requirement-id` is the same requirement.

```yaml
adherence:
  assessment-plans:
    - id: plan-a
      requirement-id: {reference-id: OSPS-B, entry-id: OSPS-BR-01.01}
      frequency-days: 30
      evaluation-methods:
        - {id: m-a, type: Behavioral, mode: Automated, required: true}
    - id: plan-b
      requirement-id: {reference-id: OSPS-B, entry-id: OSPS-BR-01.01}
      frequency-days: 90
      evaluation-methods:
        - {id: m-b, type: Intent, mode: Manual, required: true}
```

Add to `test/schema_test.go` in the Policy block:

```go
{"policy with two plans for one requirement", "./test-data/bad-policy-duplicate-requirement-plan.yaml", "#Policy", true, ""},
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/policy_with_two_plans_for_one_requirement' ./...`
Expected: FAIL — nothing forbids two plans per requirement.

- [ ] **Step 3: Rewrite `#AssessmentPlan`**

```cue
// AssessmentPlan defines how a specific assessment requirement is evaluated.
#AssessmentPlan: {
	id: string

	// requirement-id names the assessment requirement this plan governs, in the
	// catalog it comes from.
	"requirement-id": #EntryMapping @go(RequirementId)

	// frequency-days is how often this requirement is re-examined, in days.
	// Cadence belongs to the requirement, not to a method or a plan instance,
	// which is why requirement-id is unique across a policy's plans.
	"frequency-days": int & >0 @go(FrequencyDays)

	"evaluation-methods": [#AcceptedMethod & {type: #EvaluationMethodType}, ...#AcceptedMethod & {type: #EvaluationMethodType}] @go(EvaluationMethods)

	// evidence-expectation states what evidence this requirement needs and how
	// long a collected item stays usable. It declares no evidence types: an
	// expectation belongs to a policy and a requirement, never to an instance.
	"evidence-expectation"?: #EvidenceExpectation @go(EvidenceExpectation)

	parameters?: [#Parameter, ...#Parameter]

	// Parameter ids must be unique within their plan so an executor resolving a
	// parameter by id gets one value.
	if parameters != _|_ {
		_uniqueParameterIds: {for i, p in parameters {(p.id): i}}
	}
}

// EvidenceExpectation states what evidence a requirement needs and how long a
// collected item stays usable. valid-for-days is a different clock from
// frequency-days: how long evidence stays probative, not how often the
// requirement is re-examined.
#EvidenceExpectation: {
	description:      string
	"valid-for-days": int & >0 @go(ValidForDays)
}
```

Preserve `_uniqueParameterIds` from Task 3's merge.

- [ ] **Step 4: Add the uniqueness rule to `#Policy`**

Beside the `_uniquePlanIds` block Task 3 brought in:

```cue
	// One requirement has exactly one plan, so it has exactly one cadence with
	// nothing to disagree with. Several methods inside one plan cover
	// multi-source evaluation.
	if adherence."assessment-plans" != _|_ {
		_uniquePlannedRequirements: {
			for i, p in adherence."assessment-plans" {
				"\(p."requirement-id"."reference-id")/\(p."requirement-id"."entry-id")": i
			}
		}
	}
```

- [ ] **Step 5: Fix the two doc comments (21, 24)**

In `#AcceptedMethod`:

```cue
	// required means an assessment run under this plan MUST use this method.
	required: *false | bool @gemara(default=false)
```

In `#Constraint`, replace the "to enable Layer 5/6 tracking" claim:

```cue
// Constraint is a prescriptive requirement that becomes part of the requirement
// produced by resolution, so an assessment run under a plan is run against the
// constrained requirement. It is not tracked separately in any log.
```

- [ ] **Step 6: Update the two policy fixtures**

`good-policy.yaml` and `good-security-policy.yml` have no assessment plans today (that is thread 22a). Add `requirement-id` as an `#EntryMapping` and `frequency-days` wherever a plan appears after this change, and add `mapping-references` entries for any newly cited catalog so Task 1's rule holds.

- [ ] **Step 7: Add the Policy reference-id site deferred from Task 1**

Now that `requirement-id` is an `#EntryMapping`, add it to `#Policy`'s `_refValidation` block:

```cue
	for i, p in adherence."assessment-plans" {
		_refValidation: "plans-\(i)": _declaredRefIds & list.Contains(p."requirement-id"."reference-id")
	}
```

- [ ] **Step 8: Run the gates and commit**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

`breaking-check` will report `frequency` removed and `requirement-id` changed type. `policy.cue` is experimental, so both are minors — add `.oasdiff-allow` entries naming them.

```bash
git add policy.cue .oasdiff-allow test/
git commit -m "$(cat <<'EOF'
feat(policy): make assessment plan commitments checkable

frequency was free prose that could not be compared with timestamps,
evidence-requirements was prose with no consumer, and requirement-id was
a bare id that guessed at its catalog. Replace them with frequency-days,
an EvidenceExpectation carrying what is needed plus valid-for-days, and
an EntryMapping. Cadence belongs to the requirement, so requirement-id
is now unique across a policy's plans.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 6: Record — `#ExecutionFacts` (threads 21, 22c, 23)

An `#AssessmentLog` can cite a plan and record nothing about executing it. Running a plan means selecting one of its methods, so the log records that selection and what came with it.

**Files:**
- Modify: `evaluationlog.cue` (`#AssessmentLog`, new `#ExecutionFacts`), `primitives.cue` (new `#ParameterValue`)
- Create: `test/test-data/bad-evaluation-log-plan-without-execution.yaml`, `test/test-data/bad-evaluation-log-execution-without-plan.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: `#Actor` (entities.cue:33), `_uniqueMethodIds` from Task 3 — which is what makes a bare `method-id` unambiguous, since method ids are unique across a whole policy.
- Produces: `#ExecutionFacts: {"method-id": string, executor?: #Actor, parameters?: [#ParameterValue, ...]}`; `#ParameterValue: {id: string, value: string}`. Task 7 keys identity on `execution."method-id"`.

- [ ] **Step 1: Write the two failing fixtures**

`bad-evaluation-log-plan-without-execution.yaml` — an assessment with `plan` set and no `execution`.
`bad-evaluation-log-execution-without-plan.yaml` — an assessment with `execution` set and no `plan`.

Both start from `good-evaluation-log-unstarted.yaml`. Add to `test/schema_test.go`:

```go
{"assessment citing a plan without execution facts", "./test-data/bad-evaluation-log-plan-without-execution.yaml", "#EvaluationLog", true, ""},
{"assessment with execution facts but no plan", "./test-data/bad-evaluation-log-execution-without-plan.yaml", "#EvaluationLog", true, ""},
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/assessment_' ./...`
Expected: both FAIL with "expected validation error, got nil".

- [ ] **Step 3: Add `#ParameterValue` to `primitives.cue`**

It is referenced by stable `evaluationlog.cue` while its counterpart `#Parameter` lives in experimental `policy.cue`, so it cannot live in either.

```cue
// ParameterValue records the value an execution used for a plan parameter.
// The id names a #Parameter declared by the plan that was executed.
#ParameterValue: {
	id:    string
	value: string
}
```

- [ ] **Step 4: Add `#ExecutionFacts` and the rule**

In `evaluationlog.cue`:

```cue
// ExecutionFacts records what a run actually used under its plan: which of the
// plan's methods ran, who ran it when that differs from the run's executor, and
// the parameter values used. A log cites a plan to say what was supposed to
// happen; these are what did.
#ExecutionFacts: {
	// method-id names the accepted method from the cited plan that this
	// assessment executed. Method ids are unique across a policy, so no further
	// qualifier is needed.
	"method-id": string @go(MethodId)

	// executor is the actor that performed this assessment, recorded only when
	// it differs from the run's executor.
	executor?: #Actor

	// parameters are the values this execution used for the plan's parameters.
	parameters?: [#ParameterValue, ...#ParameterValue]
}
```

And inside `#AssessmentLog`:

```cue
	
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/assessment_' ./...`
Expected: PASS.

- [ ] **Step 6: Run the gates and commit**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

Confirm `breaking-check` reports only additions. `evaluationlog.cue` is stable — if it reports a removal or rename, stop and fix.

```bash
git add evaluationlog.cue primitives.cue test/
git commit -m "$(cat <<'EOF'
feat(evaluationlog): record what a run used under its plan

An assessment could cite a plan and record nothing about executing it,
so no one could check the run against the plan. Add ExecutionFacts with
the selected method, the executor when it differs from the run's, and
the parameter values used. A cited plan now obliges the record, and the
values cannot appear without a plan to validate them against.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 7: Widen assessment identity to the 4-tuple (D8)

Task 3 brought in `_uniqueAssessedRequirements`, which allows one assessment per requirement per control evaluation. A plan may declare two `required` methods, each producing its own assessment — so that rule makes a method that ran and passed unrepresentable. Identity becomes (evaluation log, requirement, `plan-id`, `method-id`).

**Files:**
- Modify: `evaluationlog.cue` (`_uniqueAssessedRequirements`)
- Create: `test/test-data/good-evaluation-log-two-methods.yaml`, `test/test-data/bad-evaluation-log-duplicate-identity.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: `execution."method-id"` from Task 6; `plan` (evaluationlog.cue:49).
- Produces: the uniqueness rule Task 9's coverage relies on to address an assessment without a minted id.

- [ ] **Step 1: Write the failing positive fixture**

`good-evaluation-log-two-methods.yaml` — one control evaluation with **two** assessments for the same requirement, same plan, different `execution.method-id`. This is legitimate data that the current rule rejects.

```go
{"two required methods assessed for one requirement", "./test-data/good-evaluation-log-two-methods.yaml", "#EvaluationLog", false, ""},
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/two_required_methods' ./...`
Expected: FAIL with "unexpected validation error" — `_uniqueAssessedRequirements` rejects the second assessment.

- [ ] **Step 3: Write the failing negative fixture**

`bad-evaluation-log-duplicate-identity.yaml` — two assessments with the same requirement, same plan **and** the same `execution.method-id`.

```go
{"two assessments with the same identity", "./test-data/bad-evaluation-log-duplicate-identity.yaml", "#EvaluationLog", true, ""},
```

- [ ] **Step 4: Widen the rule**

Replace `_uniqueAssessedRequirements` in `#EvaluationLog`:

```cue
	// An AssessmentLog has no id: it is one method's execution against one
	// requirement under one plan, so (this log, requirement, plan-id, method-id)
	// is its key. A plan may require two methods, each producing its own
	// assessment, so the method is part of the key. An assessment with no plan
	// has no method to select and falls back to the requirement alone.
	for i, e in evaluations {
		_uniqueAssessments: "\(i)": {
			for j, a in e."assessment-logs" {
				let _plan = [if a.plan != _|_ {a.plan."entry-id"}, "-"][0]
				let _method = [if a.execution != _|_ {a.execution."method-id"}, "-"][0]
				"\(a.requirement."entry-id")/\(_plan)/\(_method)": j
			}
		}
	}
```

If the `[if cond {v}, fallback][0]` idiom is rejected by CUE, use two guarded comprehensions instead — one branch for assessments with a plan, one for those without — and keep both writing into `_uniqueAssessments`.

- [ ] **Step 5: Run both tests to verify they pass**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/two_' ./...`
Expected: the positive fixture PASSES, the negative fixture PASSES (i.e. is correctly rejected).

- [ ] **Step 6: Run the gates and commit**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

```bash
git add evaluationlog.cue test/
git commit -m "$(cat <<'EOF'
fix(evaluationlog): key an assessment by requirement, plan and method

An assessment is one method's execution, and a plan may declare two
required methods, so one assessment per requirement made a method that
ran and passed impossible to record. Widen the uniqueness rule to the
full key, falling back to the requirement alone when no plan is cited.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 8: Record — the effective requirement set (D2, 14c)

A Policy never lists what it requires; the set is computed. The log that did the computing writes down what it got, giving the audit a denominator it did not derive itself.

**Files:**
- Modify: `evaluationlog.cue` (`#EvaluationLog`, new `#EffectiveRequirement`)
- Create: `test/test-data/bad-evaluation-log-policy-without-requirements.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: `#ArtifactMapping`, `#EntryMapping`.
- Produces: `#EffectiveRequirement: {requirement: #EntryMapping, "plan-id"?: string}` and `#EvaluationLog.policy` / `."effective-requirements"`. Task 9's coverage has one entry per element of this list.

- [ ] **Step 1: Write the failing fixture**

`bad-evaluation-log-policy-without-requirements.yaml` — an EvaluationLog with `policy` set and no `effective-requirements`.

```go
{"evaluation log citing a policy without its requirement set", "./test-data/bad-evaluation-log-policy-without-requirements.yaml", "#EvaluationLog", true, ""},
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/evaluation_log_citing_a_policy' ./...`
Expected: FAIL — neither field exists yet, so the extra keys are simply rejected or ignored; confirm the failure message before proceeding.

- [ ] **Step 3: Add the fields**

```cue
// EffectiveRequirement is one entry of the requirement set a run resolved its
// policy to. It records identity and resolution outcome only — no requirement
// text, no constraint prose, no inlined plan. A reader who wants the words
// follows the pointer to the artifact at its pinned version.
#EffectiveRequirement: {
	requirement: #EntryMapping

	// plan-id names the plan bound to this requirement, if any.
	"plan-id"?: string @go(PlanId)
}
```

In `#EvaluationLog`:

> We don't need this. Evaluation is about control implementation. It points to catalogs. Audit points to policy.

```cue
	// policy is the policy this run resolved, at its pinned version.
	policy?: #ArtifactMapping

	// effective-requirements is what this run determined the policy requires.
	// A run that cites a policy MUST record the set it resolved, so an audit
	// has a denominator it did not compute itself.
	"effective-requirements"?: [#EffectiveRequirement, ...#EffectiveRequirement] @go(EffectiveRequirements)

	if policy != _|_ {
		"effective-requirements": [#EffectiveRequirement, ...#EffectiveRequirement]
	}
```

- [ ] **Step 4: Extend the reference check to the new sites**

Add `policy."reference-id"` and each `"effective-requirements"[]."requirement"."reference-id"` to `#EvaluationLog`'s `_refValidation` block from Task 1.

- [ ] **Step 5: Run the test to verify it passes, then the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`
Expected: additions only on the stable file.

- [ ] **Step 6: Commit**

```bash
git add evaluationlog.cue test/
git commit -m "$(cat <<'EOF'
feat(evaluationlog): record the effective requirement set a run resolved

A Policy never lists what it requires — the set is computed, and six of
the eight resolution steps are undefined, so two resolvers can disagree
silently. The log that resolved the policy now writes down what it got,
as identity and plan binding only. An audit then has a denominator it
did not compute itself, and resolver disagreement becomes reportable.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 9: Check — `#AuditLog` coverage (D1, D3, D4, threads 16, 17, 28)

The audit's yardstick becomes one versioned Policy, and the audit gains a place to state what it verified.

**Files:**
- Modify: `auditlog.cue` (`#AuditLog`, `#AuditResult`, new `#RequirementCoverage`, `#CheckOutcome`)
- Modify: `test/test-data/good-audit-log.yaml`, `test/test-data/bad-audit-log.yaml`, `test/test-data/bad-audit-log-undeclared-criteria.yaml`
- Create: `test/test-data/bad-audit-log-coverage-missing-assessment-present.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: `#EntryMapping`, `#EvidenceMapping`, `#ArtifactMapping`, and `#EffectiveRequirement` from Task 8 as the set coverage must mirror.
- Produces: `#RequirementCoverage`, `#CheckOutcome`. Nothing later depends on these.

- [ ] **Step 1: Write the failing fixture**

`bad-audit-log-coverage-missing-assessment-present.yaml` — a coverage entry with `plan-bound` but no `assessment-present`.

```go
{"coverage entry without assessment-present", "./test-data/bad-audit-log-coverage-missing-assessment-present.yaml", "#AuditLog", true, ""},
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/coverage_entry' ./...`
Expected: FAIL — no coverage type exists yet.

- [ ] **Step 3: Replace `criteria` with `policy` and add coverage**

```cue
#AuditLog: {
	#Log
	metadata: type: "AuditLog"

	owner?:  #RACI @go(Owner)
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
}

// RequirementCoverage records what an audit checked for one requirement. Each
// check is a named field so that a check never made is visible as a missing
// field rather than an absent list entry.
#RequirementCoverage: {
	// requirement names an entry of the evaluation log's effective-requirements.
	requirement: #EntryMapping

	// assessments are the assessments this audit read, if any.
	assessments?: [#EvidenceMapping, ...#EvidenceMapping]

	// assessment-present records whether the requirement was assessed at all.
	// It is required: it is what makes coverage complete rather than selective.
	"assessment-present": #CheckOutcome @go(AssessmentPresent)

	"plan-bound"?:         #CheckOutcome @go(PlanBound)
	"method-allowed"?:     #CheckOutcome @go(MethodAllowed)
	"executor-matched"?:   #CheckOutcome @go(ExecutorMatched)
	"parameters-matched"?: #CheckOutcome @go(ParametersMatched)
	"frequency-met"?:      #CheckOutcome @go(FrequencyMet)
	"evidence-present"?:   #CheckOutcome @go(EvidencePresent)
	"evidence-fresh"?:     #CheckOutcome @go(EvidenceFresh)

	// required-methods records one outcome per method the plan requires.
	"required-methods"?: [...{"method-id": string, outcome: #CheckOutcome}] @go(RequiredMethods)

	// evidence-fresh only means something once evidence was found.
	if "evidence-fresh" != _|_ {"evidence-present": #CheckOutcome}
}

// CheckOutcome is the result of one audit check.
#CheckOutcome: {
	outcome:  "Pass" | "Fail" | "Not Applicable" | "Undetermined"
	message?: string
}
```

- [ ] **Step 4: Point `#AuditResult` into coverage**

Replace `"criteria-reference": #MultiEntryMapping` (auditlog.cue:57) with a pointer to the requirement whose failed check prompted the result:

```cue
	// requirement names the coverage entry this result was prompted by, so an
	// opinion and the mechanical check that triggered it agree.
	requirement?: #EntryMapping
```

Remove the `_criteriaValidation` block (auditlog.cue:28-36) and replace it with a check that each result's `requirement` names a coverage entry:

```cue
	let _coveredRequirements = [for c in coverage {"\(c.requirement."reference-id")/\(c.requirement."entry-id")"}]
	for i, r in results if r.requirement != _|_ {
		_resultCoverage: "\(i)": _coveredRequirements & list.Contains("\(r.requirement."reference-id")/\(r.requirement."entry-id")")
	}
```

- [ ] **Step 5: Update the audit fixtures**

`good-audit-log.yaml`: replace the `criteria` list with a single `policy` mapping to `security-policy`, add a `coverage` list with one entry per requirement its results mention, each carrying `assessment-present`, and change each result's `criteria-reference` to `requirement`. Fix the `entry-id` that points at nothing (29e) while here.

`bad-audit-log-undeclared-criteria.yaml`: it exercised the old `criteria` rule. Repurpose it to cite a `policy` whose `reference-id` is undeclared, which Task 1's rule rejects, and rename it `bad-audit-log-undeclared-policy.yaml`; update the row in `schema_test.go`.

- [ ] **Step 6: Run the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

`breaking-check` will report `criteria` and `criteria-reference` removed. `auditlog.cue` is experimental, so both are minors — add `.oasdiff-allow` entries naming them and the reason.

- [ ] **Step 7: Commit**

```bash
git add auditlog.cue .oasdiff-allow test/
git commit -m "$(cat <<'EOF'
feat(auditlog): audit against one policy, with per-requirement coverage

criteria accepted any artifact as a yardstick, and the only output was
narrative opinion, so an audit could not say that two requirements were
never assessed. Replace criteria with a single versioned policy and add
coverage: one entry per effective requirement, each check a named field
so a check never made is visible as a missing field. Risk tolerance gets
the same treatment instead of being smuggled in as a result.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 10: Result semantics doc comments (threads 26a, 26b)

Coverage reports `Pass`/`Fail` per requirement, which is meaningless until "Passed" has a written derivation. The SDK's rule is canonical but undocumented, and one fixture disagrees with it.

**Files:**
- Modify: `evaluationlog.cue` (`#Result`, `#AssessmentLog.result`, `#AssessmentLog."steps-executed"`)
- Modify: `test/test-data/good-evaluation-log-unstarted.yaml:21`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: doc-only, plus one CUE bound on `steps-executed`.

- [ ] **Step 1: Write the failing fixture**

`bad-evaluation-log-steps-executed-overflow.yaml` — an assessment whose `steps-executed` exceeds `len(steps)`.

```go
{"steps-executed greater than steps", "./test-data/bad-evaluation-log-steps-executed-overflow.yaml", "#EvaluationLog", true, ""},
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd test && go test -count=1 -run 'TestSchemaValidation/steps-executed' ./...`
Expected: FAIL — nothing bounds the field.

- [ ] **Step 3: Document the precedence and bound the field**

```cue
// Result is an assessment outcome. Aggregation at every rollup (assessment →
// control evaluation → log) follows the precedence
// Failed > Unknown > Needs Review > Passed > Not Applicable, and Not Run never
// overwrites another value.
#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown" @go(-)
```

On `#AssessmentLog.result`, replace the "matching the result of the last step that was run" wording — it contradicts the precedence, since the SDK halts only on `Failed` and `Not Applicable`:

```cue
	// result is the aggregate outcome of this assessment's steps, by the
	// precedence documented on #Result.
	result: #Result
```

And bound the counter:

```cue
	// steps-executed is how many steps ran; it never exceeds the number of
	// steps, and the failing step is the last one executed.
	"steps-executed"?: int & >=0 & <=len(steps) @go(StepsExecuted)
```

- [ ] **Step 4: Fix the fixture that disagrees with the canonical rule**

`good-evaluation-log-unstarted.yaml:21` rolls `OSPS-AC-01` up to `Needs Review` while one of its assessments is `Unknown` (:39). Under the precedence, `Unknown` wins. Change line 21 to `Unknown`.

- [ ] **Step 5: Run the gates and commit**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`

```bash
git add evaluationlog.cue test/
git commit -m "$(cat <<'EOF'
docs(evaluationlog): write down the canonical result precedence

The SDK's aggregation rule is canonical but appeared nowhere a producer
would look, and AssessmentLog.result's comment contradicted it. State
the precedence on #Result, bound steps-executed by the number of steps,
and fix the fixture that rolled an Unknown assessment up to Needs
Review.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 11: The scenario fixture set (thread 22a)

Nothing in the repo exercises an assessment plan, so none of Tasks 5–9 is exercised end to end by real data. One coherent set threads the four types by real ids.

**Files:**
- Create: `test/test-data/scenario/policy.yaml`, `test/test-data/scenario/evaluation-log.yaml`, `test/test-data/scenario/audit-log.yaml`
- Modify: `test/test-data/good-enforcement-log.yaml`
- Test: `test/schema_test.go`

**Interfaces:**
- Consumes: every type this plan changed.
- Produces: the only end-to-end proof that the chain holds.

- [ ] **Step 1: Write the Policy**

`test/test-data/scenario/policy.yaml` — a Policy whose `adherence.assessment-plans` has one plan with **two required methods**, parameters with `accepted-values`, `frequency-days`, and an `evidence-expectation`:

```yaml
adherence:
  assessment-plans:
    - id: plan-branch-protection
      requirement-id: {reference-id: OSPS-B, entry-id: OSPS-BR-01.01}
      frequency-days: 30
      evaluation-methods:
        - {id: m-intent, type: Intent, mode: Manual, required: true,
           description: "Documented branch protection policy review"}
        - {id: m-behavioral, type: Behavioral, mode: Automated, required: true,
           description: "Probe repository settings via the API"}
      parameters:
        - id: min-approvals
          label: Minimum approvals
          description: Reviewers required before merge
          accepted-values: ["1", "2"]
      evidence-expectation:
        description: "Repository settings snapshot and the reviewer's sign-off"
        valid-for-days: 90
```

Declare `OSPS-B` in `metadata.mapping-references` and give the file a `metadata.version`.

- [ ] **Step 2: Write the EvaluationLog**

`test/test-data/scenario/evaluation-log.yaml` — cites the policy, carries `effective-requirements`, and holds **two** assessments for `OSPS-BR-01.01`, one per method, each with its own `execution`, result, timing and evidence:

```yaml
policy: {reference-id: acme-policy}
effective-requirements:
  - requirement: {reference-id: OSPS-B, entry-id: OSPS-BR-01.01}
    plan-id: plan-branch-protection
# ...
        assessment-logs:
          - requirement: {reference-id: OSPS-B, entry-id: OSPS-BR-01.01}
            plan: {reference-id: acme-policy, entry-id: plan-branch-protection}
            execution:
              method-id: m-behavioral
              parameters:
                - {id: min-approvals, value: "2"}
            # ...result, message, applicability, steps, start, end, evidence
```

Declare `acme-policy` and `OSPS-B` in its `mapping-references`.

- [ ] **Step 3: Write the AuditLog**

`test/test-data/scenario/audit-log.yaml` — `policy` pointing at the same policy, and `coverage` with one entry for `OSPS-BR-01.01` carrying `assessment-present`, `plan-bound`, `method-allowed`, `parameters-matched`, `frequency-met`, `evidence-present`, `evidence-fresh`, and a `required-methods` entry per method.

- [ ] **Step 4: Chain the enforcement log**

Point `good-enforcement-log.yaml`'s `AssessmentFinding.log` at the scenario evaluation log's real ids, closing #499's dangling pointer with data that resolves.

- [ ] **Step 5: Register all three**

```go
// Scenario — the full chain, threaded by real ids
{"scenario policy", "./test-data/scenario/policy.yaml", "#Policy", false, ""},
{"scenario evaluation log", "./test-data/scenario/evaluation-log.yaml", "#EvaluationLog", false, ""},
{"scenario audit log", "./test-data/scenario/audit-log.yaml", "#AuditLog", false, ""},
```

- [ ] **Step 6: Run the gates**

Run: `cd test && go test -count=1 ./... && cd .. && make cuefmtcheck && make lintcue && make genopenapi && make breaking-check`
Expected: all three validate. If the two-method evaluation log is rejected, Task 7's key is wrong — fix there, not here.

- [ ] **Step 7: Commit**

```bash
git add test/
git commit -m "$(cat <<'EOF'
test(fixtures): add the policy to audit scenario

No fixture exercised an assessment plan, so nothing proved the chain
holds end to end. Add a policy whose plan requires two methods, an
evaluation log recording an execution per method with the requirement
set it resolved, and an audit log covering that requirement with named
checks. The enforcement log now points at ids that resolve.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

### Task 12: Tool-spec text for the cross-document rules

CUE cannot see across documents, so the rules whose two halves live in different artifacts have to be written for tool developers. Without this, the design's central rule — coverage completeness — exists nowhere.

**Files:**
- Create: `docs/tool-spec/evidence-evaluation-audit.md`

**Interfaces:**
- Consumes: the field names from Tasks 5–9. Every name in this document must match the schema exactly.

- [ ] **Step 1: Write the document**

Cover each rule with what it checks and what a tool reports:

1. **Coverage completeness** — exactly one coverage entry per element of the evaluation log's `effective-requirements`, none extra. The rule this design exists for.
2. **Check conditionality** — if the plan defines the thing a check covers, the coverage entry reports that check; a missing field means not checked.
3. **Version pins** — a `#MappingReference.version` equals the fetched artifact's `metadata.version`; report a mismatch.
4. **Cadence** — for each method the plan marks `required`, consecutive scheduled runs (by `#Log.run.trigger` and each assessment's `start`) fall within `frequency-days`.
5. **Parameters** — every recorded `execution.parameters[].id` is declared by the plan, every value is within that parameter's `accepted-values`, and every plan parameter was recorded.
6. **Evidence** — the requirement's assessments *collectively* carry at least one `#Evidence` whose `collected-at` is within `valid-for-days`.
7. **Executor** — `execution.executor` matches the plan's method executor **by authoritative identifier** (purl, ARN, SPIFFE), never by the internal short-name `id`.
8. **Spot check** — fetch `source.download-url`, recompute `digest`, report verified / integrity-failure / unverifiable, never a boolean. Referenced evidence with no digest is reported as not spot-checkable.
9. **Required methods** — every method the plan marks `required` has an assessment.

- [ ] **Step 2: Cross-check every field name against the schema**

Run: `grep -o '`[a-z-]*`' docs/tool-spec/evidence-evaluation-audit.md | sort -u` and confirm each name appears in the CUE.

- [ ] **Step 3: Commit**

```bash
git add docs/tool-spec/
git commit -m "$(cat <<'EOF'
docs(tool-spec): specify the cross-document chain rules

CUE holds what one document can check about itself; coverage
completeness, cadence, parameter validity, evidence freshness, executor
matching and the digest spot check all span documents. Write them down
so two implementations agree on whether a requirement passed.

Signed-off-by: Jennifer Power <barnabei.jennifer@gmail.com>
Assisted-by: Claude (model: claude-opus-5)
EOF
)"
```

---

## Self-Review

**Spec coverage.** §4.1 → Task 5. §4.2 → Tasks 6, 7, 8. §4.3 → Task 4. §4.4 → Task 9. §4.5 → Tasks 5–9 collectively, verified by Task 11. §5 CUE rules → Tasks 1, 5, 6, 7, 8, 9, 10. §5 tool-spec rules → Task 12. §6 repairs → Tasks 1, 2, 9, 10; scenario → Task 11; negative fixtures → one per rule-introducing task. §7 → the `.oasdiff-allow` steps in Tasks 2, 5, 9. §8 sequencing → Tasks 1–12 in order. Prerequisites → Tasks 1, 2, 3.

**Known gaps, stated rather than hidden:**
- The spec's 1a rule says "every artifact root". Task 1 covers `#Policy`, `#EvaluationLog`, `#AuditLog` and `#ControlCatalog` — the roots this chain uses and the ones whose fixtures are provably broken. `#GuidanceCatalog`, `#ThreatCatalog`, `#RiskCatalog`, `#CapabilityCatalog`, `#PrincipleCatalog` and `#MappingDocument` are left to the mapping primitives work stream, which the spec scopes out.
- Two CUE constructs are used that this repo has no precedent for: the chained `if … for …` comprehension clause (Task 1) and the `[if cond {v}, fallback][0]` default idiom (Task 7). Each step names the fallback formulation to use if CUE rejects it. Both surface immediately in `make lintcue`.

**Type consistency.** `#ExecutionFacts` (Task 6) is read by Task 7's key and Task 9's checks under the same field names (`method-id`, `executor`, `parameters`). `#EffectiveRequirement` (Task 8) is mirrored by `#RequirementCoverage.requirement` (Task 9). `#CheckOutcome` is defined once in Task 9 and used by `risk-tolerance-met` and every coverage field. `#ParameterValue` is defined in `primitives.cue` (Task 6) and referenced only by `#ExecutionFacts`.
