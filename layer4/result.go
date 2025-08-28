package layer4

import "encoding/json"

type Result struct {
	Status Status `json:"status"`

	Message string `json:"message"`
}

// Status is an enum representing the result of a control evaluation
// This is designed to restrict the possible result values to a set of known states
type Status int

const (
	NotRun Status = iota
	Passed
	Failed
	NeedsReview
	NotApplicable
	Unknown
)

var toString = map[Status]string{
	NotRun:        "Not Run",
	Passed:        "Passed",
	Failed:        "Failed",
	NeedsReview:   "Needs Review",
	NotApplicable: "Not Applicable",
	Unknown:       "Unknown",
}

func (s Status) String() string {
	return toString[s]
}

// MarshalYAML ensures that Status is serialized as a string in YAML
func (s Status) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

// MarshalJSON ensures that Status is serialized as a string in JSON
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UpdateAggregateStatus compares the current result with the new result and returns the most severe of the two.
func UpdateAggregateStatus(previous Status, new Status) Status {
	if new == NotRun {
		// Not Run should not overwrite anything
		// Failed should not be overwritten by anything
		// Failed should overwrite anything
		return previous
	}

	if previous == Failed || new == Failed {
		// Failed should not be overwritten by anything
		// Failed should overwrite anything
		return Failed
	}

	if previous == Unknown || new == Unknown {
		// If the current or past result is Unknown, it should not be overwritten by NeedsReview or Passed.
		return Unknown
	}

	if previous == NeedsReview || new == NeedsReview {
		// NeedsReview should not be overwritten by Passed
		// NeedsReview should overwrite Passed
		return NeedsReview
	}
	return Passed
}
