package layer4

import "encoding/json"

// ExecutorType specifies whether an executor is automated or manual.
type ExecutorType string

const (
	// Automated indicates the executor is a tool or script that runs without human intervention.
	Automated ExecutorType = "Automated"

	// Manual indicates the executor requires human review or judgment.
	Manual ExecutorType = "Manual"
)

func (e ExecutorType) String() string {
	return string(e)
}

// MarshalJSON ensures that ExecutorType is serialized as a string in JSON
func (e ExecutorType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(e))
}

// MarshalYAML ensures that ExecutorType is serialized as a string in YAML
func (e ExecutorType) MarshalYAML() (interface{}, error) {
	return string(e), nil
}
