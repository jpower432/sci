package layer4

import (
	"encoding/json"
)

// Method is an enum representing the method used to determine the assessment procedure result.
// This is designed to restrict the possible method values to a set of known types.
type Method int

const (
	UnknownMethod Method = iota
	// AutomatedMethod represents an automated testing assessment method
	AutomatedMethod
	// ManualMethod represents an assessment method that requires
	// inspection done by a human
	ManualMethod
)

// methodToString maps Method values to their string representations.
var methodToString = map[Method]string{
	ManualMethod:    "Manual",
	AutomatedMethod: "Automated",
	UnknownMethod:   "Unknown",
}

func (m Method) String() string {
	return methodToString[m]
}

// MarshalYAML ensures that Method is serialized as a string in YAML
func (m Method) MarshalYAML() (interface{}, error) {
	return m.String(), nil
}

// MarshalJSON ensures that Method is serialized as a string in JSON
func (m Method) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}
