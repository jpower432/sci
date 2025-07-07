package layer4

// TODO: This file is currently manually maintained rather than generated due to issues with `cue exp gengotypes`
type AssessmentMethod struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	ID               string            `json:"id,omitempty"`
	Run              bool              `json:"run"`
	RemediationGuide string            `json:"remediation_guide,omitempty"`
	Documentation    string            `json:"documentation,omitempty"`
	Result           *AssessmentResult `json:"result"`
	Executor         MethodExecutor    `json:"-"`
}

// MethodExecutor is a function type that inspects the provided payload and returns the result of the assessment.
// The payload is the data/evidence that the assessment will be run against.
type MethodExecutor func(payload interface{}, c map[string]*Change) (AssessmentResult, error)

// RunMethod executes the assessment method using the provided payload and changes.
// It returns the result of the assessment and any error encountered during execution.
// The payload is the data/evidence that the assessment will be run against.
func (a *AssessmentMethod) RunMethod(payload interface{}, changes map[string]*Change) (AssessmentResult, error) {
	result, err := a.Executor(payload, changes)
	a.Result = &result
	a.Run = true
	return result, err
}
