# A complete chain, at the scale of a real baseline

Illustrative sample data, not a real audit of anything. It exists to answer what the
smaller fixtures cannot: does the model hold when it covers every requirement of an
actual baseline, rather than four hand-picked ones?

- `policy.yaml` — an adherence policy adopting the
  [OpenSSF Project Security Baseline](https://baseline.openssf.org) at Maturity Level 2
  and a supply-chain guidance document, binding plans to twelve requirements,
  excluding three, and declaring one fallback method for everything else.
- `evaluation-log.yaml` — the scan: 32 control evaluations over 44 assessments, each
  naming the executor that produced it and the method it ran.
- `enforcement-log.yaml` — what the release gate did: four releases blocked, one
  requirement remediated, one waived under an exception, and a `Clear` re-run.
- `audit-log.yaml` — the audit conducted against all of it.

## What it demonstrates

**Completeness.** One verification per requirement the policy governs — 56 of the
catalog's 59, since the policy excludes three — which is what makes an audit
systematic rather than selective. The whole audit is 30 KiB, well under half the size
of the catalog it audits.

**Both method sources.** A plan prescribes methods for twelve requirements; the
policy's own `evaluation-methods` are the fallback for the rest. Conformance is
recorded against whichever applied, and the evaluation log names the method it ran in
either case.

**Divergence with its values.** `OSPS-AC-03.01` has a conformance check that failed —
the probe last ran five months before the window — so the entry states what was
required and what was evidenced. The 43 checks that passed state neither: restating
the prescription back at a reader is how a complete record becomes an unreadable one.

**Coverage of the mandate.** The policy adopted four guidance objectives and selected
requirements implementing three. The fourth, `SCG-PROVENANCE`, is recorded as
`implementation: Not Satisfied` — a gap no verification could show, because every
requirement the policy *did* select is satisfied.

**Response.** Two findings name the enforcement action that changed their lifecycle:
one `Resolved` by a remediation, one `Waived` under an exception.

## What the schema cannot check here

Every file validates on its own. These need two or more documents, and a tool:

1. **Completeness** — one entry per governed requirement, none for an excluded one.
2. **Effective requirement fidelity** — each `effective.text` derivable from the
   catalog at the version the policy pins.
3. **Plan and method resolution** — every `plan-id` names a plan governing that
   requirement, and every `method-id` is one the policy prescribes, from the plan or
   from the fallback list.
4. **Citations resolve** — every `assessments` entry names an assessment the scan
   contains, and every finding `response` names an action the gate log contains.
5. **Synthesis** — every outcome or conformance check that is neither Satisfied nor
   Not Applicable is named by a finding.

All five pass against these files.
