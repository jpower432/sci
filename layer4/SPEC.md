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

**Description**: Uses trust scores from executor mappings to compute a weighted average of results, giving more weight to executors with higher trust scores. This allows you to prioritize results from more reliable or accurate executors when multiple executors provide conflicting results.

**Use Case**: Use when different executors have varying levels of reliability or accuracy for a specific procedure, and you want their results weighted proportionally to their trustworthiness. For example:
- A well-established tool with high accuracy should have more influence than a newer experimental tool
- A tool that specializes in a specific domain should be weighted higher for assessments in that domain
- Multiple tools with different strengths can contribute proportionally to the final determination

**Trust Scores**: Each executor has a trust score on a scale of 1 to 10, where:
- **1-3**: Low trust/reliability (e.g., experimental tools, tools with known issues)
- **4-6**: Moderate trust (e.g., standard tools with typical accuracy)
- **7-9**: High trust (e.g., well-established tools, tools with proven track records)
- **10**: Maximum trust (e.g., industry-standard tools, manually verified executors)

**Finding Determination Process**:

1. **Convert results to numeric values**:
   Each result type is assigned a numeric value representing its severity/confidence:
   - `Failed = 0` (most severe - security issue detected)
   - `Unknown = 1` (indeterminate result)
   - `NeedsReview = 2` (requires human review)
   - `Passed = 3` (no issues found)
   - `NotApplicable = 4` (not relevant to context)
   - `NotRun` = excluded from calculation (no result available)

2. **Calculate weighted average**:
   For each executor with a valid result (not NotRun), multiply its result value by its trust score, then sum all products and divide by the sum of all trust scores:
   ```
   weighted_sum = ?(result_value ? trust_score)
   total_weight = ?(trust_score)
   weighted_average = weighted_sum / total_weight
   ```
   
   This formula ensures that executors with higher trust scores have proportionally more influence on the final result.

3. **Determine finding based on weighted average thresholds**:
   The weighted average is compared against thresholds to determine the final finding:
   - If `weighted_average < 0.5`: **Finding exists** (Failed)
   - If `weighted_average < 1.5`: **Finding exists** (Unknown)
   - If `weighted_average < 2.5`: **Finding exists** (NeedsReview)
   - If `weighted_average < 3.5`: **No finding** (Passed)
   - Otherwise (`weighted_average >= 3.5`): **No finding** (NotApplicable)

**Special Cases**:
- **Failed results override weights**: If ANY executor reports `Failed`, a **Failed finding** is determined immediately, regardless of weighted average. This ensures that any detected security issue is not masked by weighted averaging.
- **NotRun exclusion**: Executors with `NotRun` results are excluded from the weighted calculation (they don't contribute to weighted_sum or total_weight).
- **All NotRun**: If all executors report `NotRun`, no finding is determined.
- **Single executor**: If only one executor has a valid result, its result becomes the final determination (weighted average equals that executor's result value).

**Example 1: Basic Weighted Average**:
- Executor A (trust-score: 7): Passed (value: 3)
- Executor B (trust-score: 9): NeedsReview (value: 2)
- Calculation:
  - weighted_sum = (3 ? 7) + (2 ? 9) = 21 + 18 = 39
  - total_weight = 7 + 9 = 16
  - weighted_average = 39 / 16 = 2.4375
- **Finding**: NeedsReview finding determined (since 2.4375 < 2.5, but >= 1.5)

**Example 2: Failed Override**:
- Executor A (trust-score: 10): Passed (value: 3)
- Executor B (trust-score: 5): Failed (value: 0)
- Even though Executor A has a higher trust score and reported Passed, **Finding**: Failed finding determined immediately because Executor B reported Failed.

**Example 3: High Trust Dominance**:
- Executor A (trust-score: 9): Passed (value: 3)
- Executor B (trust-score: 2): NeedsReview (value: 2)
- Calculation:
  - weighted_sum = (3 ? 9) + (2 ? 2) = 27 + 4 = 31
  - total_weight = 9 + 2 = 11
  - weighted_average = 31 / 11 = 2.818
- **Finding**: No finding (Passed) - The high-trust executor's Passed result dominates, resulting in weighted_average of 2.818, which is > 2.5 and < 3.5, indicating Passed.

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
