# A complete audit, at the scale of a real baseline

Illustrative sample data, not a real audit of anything. It exists to answer one
question the smaller fixtures cannot: what does a Gemara audit look like when it
covers every requirement of an actual baseline, rather than four hand-picked ones?

- `policy.yaml` — an adherence policy adopting the
  [OpenSSF Project Security Baseline](https://baseline.openssf.org) at Maturity
  Level 2, binding assessment plans to twelve requirements and excluding three.
- `evaluation-log.yaml` — the scan that evaluated it: 32 control evaluations over 44
  assessments, four requirements being assessed twice because their plan requires two
  methods.
- `audit-log.yaml` — the audit conducted against both.

Together they are a complete chain, which is what lets
`test/conformance_test.go` check the cross-document rules that CUE cannot.

## What it demonstrates

The audit carries **one verification entry per requirement the policy governs** —
56 of the catalog's 59, since the policy excludes three — which is what makes an
audit systematic rather than selective. The whole document is 31 KiB, about half
the size of the catalog it audits.

The determination mix is deliberately realistic: 32 `Compliant`, 6 `Not Compliant`,
2 `Undetermined`, and 16 `Not Applicable` where the requirement applies only at a
maturity level this policy does not adopt. Those 16 omit `evidence-present`, because a
requirement out of scope has no evidence question to answer; omitting it obliges the
outcome to be `Not Applicable`, so it cannot hide an entry left unanswered.

Three cases are worth reading in particular:

- **`OSPS-AC-03.01`** is `Undetermined`, and the entry says why without a reader
  having to cross-reference anything: the evidence is present but not fresh, and the
  probe ran under its plan with the expected executor yet missed its cadence.
- **`OSPS-AC-01.01`** is the boring pass, recorded in full. Recording it is the
  point; an audit that only lists problems cannot be shown to be complete.
- The findings include one with `status: Compliant`. A finding is not inherently
  negative — in ISO 19011 terms it is the result of evaluating evidence against the
  criteria, which may record conformity.

## What the schema cannot check here

Both files validate individually. Three rules from
[the tool specification](../../docs/tool-spec.md) need both documents plus the
catalog, and a tool must implement them:

1. **Completeness** — one entry per governed requirement, with no entry for a
   requirement the policy excludes. An early draft of this example failed exactly
   this check: it verified three excluded requirements.
2. **Effective requirement fidelity** — each `effective.text` and `applicability`
   must be derivable from the catalog at the version the policy pins.
3. **Plan and method resolution** — every `effective.plan-id` names a plan in the
   cited policy governing that requirement, and every `method-id` is one that plan
   accepts.
4. **Citations resolve** — every `assessments` entry names an assessment the scan
   actually contains. An assessment log has no id of its own, so the citation
   resolves by requirement, and where a plan requires two methods the requirement
   alone is ambiguous: those citations carry a `coordinate` naming the method.
