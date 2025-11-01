package layer4

import "encoding/json"

// ConflictRuleType specifies the type of aggregation logic used to resolve conflicts
// when multiple executors provide results for the same assessment procedure.
type ConflictRuleType string

const (
	// WeightedScore uses trust scores to compute a weighted average of results,
	// giving more weight to executors with higher trust scores. See SPEC.md for details.
	WeightedScore ConflictRuleType = "WeightedScore"

	// Strict indicates that if any executor reports a failure, the overall
	// procedure result is failed, regardless of other executor results.
	Strict ConflictRuleType = "Strict"

	// ManualOverride gives precedence to manual review executors over automated
	// executors when results conflict.
	ManualOverride ConflictRuleType = "ManualOverride"
)

func (c ConflictRuleType) String() string {
	return string(c)
}

// MarshalJSON ensures that ConflictRuleType is serialized as a string in JSON
func (c ConflictRuleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

// MarshalYAML ensures that ConflictRuleType is serialized as a string in YAML
func (c ConflictRuleType) MarshalYAML() (interface{}, error) {
	return string(c), nil
}
