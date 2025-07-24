package layer4

import (
	"encoding/json"
	"reflect"
	"runtime"
)

// Adapted from https://github.com/trumant/sci/blob/generate-go-types-from-schema/layer4/assessment_method.go

// TODO: This file is currently manually maintained rather than generated due to issues with `cue exp gengotypes`
type AssessmentMethod struct {
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Run              bool    `json:"run"`
	RemediationGuide string  `json:"remediation_guide,omitempty"`
	Documentation    string  `json:"documentation,omitempty"`
	Result           *Result `json:"result"`
	// Executor is a function type that inspects the provided targetData and returns a Result with a message.
	// The message may be an error string or other descriptive text.
	Executor MethodExecutor
}

// MethodExecutor is a function type that inspects the provided payload and returns the result of the assessment.
// The payload is the data/evidence that the assessment will be run against.
type MethodExecutor func(payload interface{}, c map[string]*Change) (Result, string)

// RunMethod executes the assessment method using the provided payload and changes.
// It returns the result of the assessment and any error encountered during execution.
// The payload is the data/evidence that the assessment will be run against.
func (a *AssessmentMethod) RunMethod(payload interface{}, changes map[string]*Change) (Result, string) {
	result, message := a.Executor(payload, changes)
	a.Result = &result
	a.Run = true
	return result, message
}

func (e MethodExecutor) String() string {
	// Get the function pointer correctly
	fn := runtime.FuncForPC(reflect.ValueOf(e).Pointer())
	if fn == nil {
		return "<unknown function>"
	}
	return fn.Name()
}

func (e MethodExecutor) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e MethodExecutor) MarshalYAML() (interface{}, error) {
	return e.String(), nil
}
