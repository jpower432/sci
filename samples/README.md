# Samples for Redesign

## Significant Decisions

### Baselines as Parent Policies

Guidance authors often create **baselines** that select and prioritize controls from standards. 

These baselines are **Layer 3 parent policies**, not Layer 1 extensions, because they:

1. **Are policies, not standards**: Baselines are policy selections of standards with operational context (severity, priority), not standards themselves
2. **Seed organization policies**: Baselines serve as parent policies that organization-specific policies inherit from
3. **Enable policy inheritance**: Organization policies can inherit from baselines and override or extend the baseline's operational context

**Architecture:**
- **Layer 1**: Pure standards/guidance with no operational context
- **Layer 3 Parent Policies**: Baselines that reference Layer 1 standards and add operational context
- **Layer 3 Child Policies**: Organization-specific policies that inherit from baselines and add organization-specific operational context (evaluators, evidence requirements, enforcement rules)

### Flatten Families

Control families were flattened from a nested structure (`ControlFamily` containing `Controls`) to a flat structure where `families` are defined at the top level and controls reference them by `family-id`.

This architectural change was driven by a practical challenge:
After controls were initially designed and numbered, they were recategorized into different families. However, changing control IDs/numbers would have caused significant confusion and broken references.

By flattening the structure and using family references:
1. **Control IDs remain stable**: Control numbering stays consistent even when families change.  
2. **Flexible categorization**: Controls can be recategorized by simply updating their `family-id` reference without changing their ID

### Move Layer 4 Evaluation Plans into Layer 3 Policy

Evaluation plans (defining *how* and *when* policies will be evaluated) were moved from Layer 4 to Layer 3 as part of the `ImplementationPlan`.

Originally, evaluation plans were separate from policies, creating a disconnect between:
- What the policy requires (Layer 3)
- How the policy will be evaluated (Layer 4)

This separation caused several issues:
1. **Policy-evaluation mismatch**: Policies could be created without clear evaluation strategies or evidence requirements.
2. **Evaluator assignment**: It was unclear which evaluators (human or automated) should assess which policies from an organization standpoint.

By moving evaluation plans into Layer 3's `ImplementationPlan`:
1. **Policy completeness**: Every policy defines its own evaluation approach, timeline, and evaluators
2. **Clear ownership**: Policies specify who/what will evaluate them (evaluators field)
3. **Implementation clarity**: The full implementation strategy (notification → evaluation → enforcement) is in one place
4. **Separation of concerns**: Layer 3 defines *plans* (what/how/when), Layer 4 contains *results* (actual findings)

**Layer Responsibilities:**
- **Layer 3 (`ImplementationPlan`)**: Defines evaluation timeline, evaluators, enforcement timeline, and methods
- **Layer 4 (`EvaluationLog`)**: Contains actual evaluation results, assessment logs, and evidence
