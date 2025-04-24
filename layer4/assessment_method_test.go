package layer4

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssessmentMethodRunMethod(t *testing.T) {
	// Create a mock payload and changes
	payload := map[string]string{"key": "value"}
	changes := map[string]*Change{}

	// Create a mock AssessmentMethod with a simple executor
	method := AssessmentMethod{
		Name:        "TestMethod",
		Description: "A test method",
		Executor: func(payload interface{}, changes map[string]*Change) (AssessmentResult, error) {
			return AssessmentResult{Status: Status("passed")}, nil
		},
	}

	// Run the method
	result, err := method.RunMethod(payload, changes)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, Status("passed"), result.Status)
	assert.True(t, method.Run)
}
func TestAssessmentMethodRunMethodWithError(t *testing.T) {
	// Create a mock payload and changes
	payload := map[string]string{"key": "value"}
	changes := map[string]*Change{}

	// Create a mock AssessmentMethod with an executor that returns an error
	method := AssessmentMethod{
		Name:        "TestMethod",
		Description: "A test method",
		Executor: func(payload interface{}, changes map[string]*Change) (AssessmentResult, error) {
			return AssessmentResult{Status: Status("error")}, fmt.Errorf("mock error")
		},
	}

	// Run the method
	result, err := method.RunMethod(payload, changes)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, Status("error"), result.Status)
	assert.True(t, method.Run)
}
