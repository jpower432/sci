# Layer 4 Finding Determination Specification

This document specifies how findings are determined in Layer 4 evaluation plans when multiple executors and procedures provide results for the same assessment requirement.

## Overview

Layer 4 evaluation plans support multiple assessment executors running assessment procedures to evaluate control requirements.
When multiple executors run the same procedure, or multiple procedures evaluate the same requirement, conflict resolution strategies
determine how their results are combined to determine if there is a finding.

## Result Types

The following result types are used in Layer 4:

- **NotRun**: The assessment was not executed
- **Passed**: The assessment passed successfully
- **Failed**: The assessment failed
- **NeedsReview**: The assessment requires manual review
- **NotApplicable**: The assessment is not applicable to the current context
- **Unknown**: The assessment result is unknown or indeterminate

## Conflict Resolution Strategies

Three conflict resolution strategies are available for determining findings from multiple executor results:

### 1. Strict

**Description**: If any executor reports a failure, a finding is determined (the procedure has failed), regardless of other executor results.

**Use Case**: Use when security-critical checks require zero tolerance for failures. Any failure from any executor (automated or manual) should be treated as a finding.

**Finding Determination**:
- **Finding exists** if ANY executor reports Failed
- **Finding exists** if ANY executor reports Unknown (and no failures)
- **Finding exists** if ANY executor reports NeedsReview (and no failures/unknown)
- **No finding** if ALL executor results are Passed
- **No finding** if ALL executor results are NotApplicable

**Example**:
- Executor A (trust-score: 7): Passed
- Executor B (trust-score: 9): Failed
- **Finding**: Failed finding determined (because Executor B reported failure)

### 2. WeightedScore

**Description**: Uses trust scores from executor mappings to compute a weighted determination of findings, giving more weight to executors with higher trust scores.

**Use Case**: Use when different executors have varying levels of reliability or accuracy for a specific procedure, and you want their results weighted accordingly.

**Trust Scores**: Each executor has a trust score on a scale of 1 to 10, where 10 is the most trusted.

**Finding Determination**:
1. Convert each result to a numeric value:
   - Failed = 0
   - Unknown = 1
   - NeedsReview = 2
   - Passed = 3
   - NotApplicable = 4
   - NotRun = excluded from calculation

2. Calculate weighted average:
   ```
   weighted_sum = ?(result_value ? trust_score)
   total_weight = ?(trust_score)
   weighted_average = weighted_sum / total_weight
   ```

3. Determine finding based on weighted average:
   - If weighted_average < 0.5: **Finding exists** (Failed)
   - If weighted_average < 1.5: **Finding exists** (Unknown)
   - If weighted_average < 2.5: **Finding exists** (NeedsReview)
   - If weighted_average < 3.5: **No finding** (Passed)
   - Otherwise: **No finding** (NotApplicable)

**Special Cases**:
- If any executor reports Failed, a **Failed finding** is determined (regardless of weights)
- If all executors report NotRun, no finding is determined
- Executors with NotRun results are excluded from the weighted calculation

**Example**:
- Executor A (trust-score: 7): Passed (value: 3)
- Executor B (trust-score: 9): NeedsReview (value: 2)
- Weighted average = (3?7 + 2?9) / (7+9) = (21 + 18) / 16 = 39/16 = 2.4375
- **Finding**: NeedsReview finding determined (since 2.4375 < 2.5)

### 3. ManualOverride

**Description**: Gives precedence to manual review executors over automated executors when determining findings from conflicting results.

**Use Case**: Use when manual reviewers should have the final say when their results differ from automated tools.

**Finding Determination**:
1. Separate results into manual and automated executor results
2. If manual executors exist:
   - If any manual executor reports Failed: **Finding exists** (Failed)
   - Else if any manual executor reports Unknown: **Finding exists** (Unknown)
   - Else if any manual executor reports NeedsReview: **Finding exists** (NeedsReview)
   - Else if all manual executors report Passed: **No finding** (Passed)
   - Else: **No finding** (NotApplicable)
3. If no manual executors exist, determine finding from automated results using severity hierarchy (Failed > Unknown > NeedsReview > Passed)

**Example**:
- Automated Executor A: Passed
- Automated Executor B: Passed
- Manual Executor C: NeedsReview
- **Finding**: NeedsReview finding determined (manual executor takes precedence)

## Examples

### Example 1: Strict Strategy at Procedure Level

**Scenario**: Branch protection check with two automated executors using Strict strategy.

```yaml
procedures:
  - id: check-branch-protection-automated
    executors:
      - id: pvtr-baseline-scanner
        trust-score: 7
      - id: openssf-scorecard
        trust-score: 9
    strategy:
      conflict-rule-type: Strict
```

**Executor Results**:
- PVTR Baseline Scanner: Passed
- OpenSSF Scorecard: Failed

**Finding Determination**: Since Scorecard reported Failed, and strategy is Strict, a **Failed finding** is determined.

### Example 2: WeightedScore Strategy

**Scenario**: Dependency manifest check with two executors using WeightedScore strategy.

```yaml
procedures:
  - id: check-dependency-manifests-automated
    executors:
      - id: pvtr-baseline-scanner
        trust-score: 9
      - id: another-tool
        trust-score: 5
    strategy:
      conflict-rule-type: WeightedScore
```

**Executor Results**:
- PVTR Baseline Scanner (trust-score: 9): Passed (value: 3)
- Another Tool (trust-score: 5): NeedsReview (value: 2)

**Finding Determination**:
- Weighted average = (3?9 + 2?5) / (9+5) = (27 + 10) / 14 = 37/14 = 2.64
- Since 2.64 > 2.5, **No finding** is determined (Passed)

### Example 3: ManualOverride Strategy

**Scenario**: MFA check with automated and manual executors using ManualOverride strategy.

```yaml
procedures:
  - id: check-mfa
    executors:
      - id: pvtr-baseline-scanner
        trust-score: 8
        type: Automated
      - id: manual-review
        trust-score: 10
        type: Manual
    strategy:
      conflict-rule-type: ManualOverride
```

**Executor Results**:
- PVTR Baseline Scanner (Automated): Passed
- Manual Review (Manual): NeedsReview

**Finding Determination**: Since ManualOverride strategy is used and manual executor reported NeedsReview, a **NeedsReview finding** is determined (manual takes precedence).

### Example 4: No Strategy Specified (Default Behavior)

**Scenario**: Multiple procedures evaluating the same requirement without an explicit strategy.

**Procedure Results**:
- Procedure A: Passed
- Procedure B: NeedsReview
- Procedure C: Passed

**Finding Determination**: Using default severity-based determination:
- NeedsReview finding is determined (NeedsReview takes precedence over Passed)

## Implementation Notes

1. **Trust Scores**: Trust scores are only used in the WeightedScore strategy. In other strategies, they are informational only.

2. **Executor Types**: Executor types (Automated vs Manual) are only used in the ManualOverride strategy.

3. **Missing Results**: If an executor is expected to run but doesn't produce a result, it's treated as NotRun, which doesn't overwrite existing results.
