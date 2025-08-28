package layer4

import "testing"

var controlEvaluationTestData = []struct {
	testName          string
	control           *ControlEvaluation
	failBeforePass    bool
	expectedResult    Status
	expectedCorrupted bool
}{
	{
		testName:          "ControlEvaluation with no Assessments",
		expectedResult:    NeedsReview,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{},
		},
	},
	{
		testName:          "ControlEvaluation with one passing AssessmentLog",
		expectedResult:    Passed,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{passingAssessmentPtr()},
		},
	},
	{
		testName:          "ControlEvaluation with one failing AssessmentLog",
		expectedResult:    Failed,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{failingAssessmentPtr()},
		},
	},
	{
		testName:          "ControlEvaluation with one NeedsReview AssessmentLog",
		expectedResult:    NeedsReview,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{needsReviewAssessmentPtr()},
		},
	},
	{
		testName:          "ControlEvaluation with one Unknown AssessmentLog",
		expectedResult:    Unknown,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{unknownAssessmentPtr()},
		},
	},
	{
		testName:          "ControlEvaluation with first NeedsReview and then Unknown AssessmentLog",
		expectedResult:    Unknown,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{
				needsReviewAssessmentPtr(),
				unknownAssessmentPtr(),
			},
		},
	},
	{
		testName:          "ControlEvaluation with first Unknown and then NeedsReview AssessmentLog",
		expectedResult:    Unknown,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{
				unknownAssessmentPtr(),
				needsReviewAssessmentPtr(),
			},
		},
	},
	{
		testName:          "ControlEvaluation with first Failed and then NeedsReview AssessmentLog",
		expectedResult:    Failed,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{
				failingAssessmentPtr(),
				needsReviewAssessmentPtr(),
			},
		},
	},
	{
		testName:          "ControlEvaluation with first Failing and then Passing AssessmentLog",
		expectedResult:    Failed,
		failBeforePass:    true,
		expectedCorrupted: false,
		control: &ControlEvaluation{
			Assessments: []*AssessmentLog{
				failingAssessmentPtr(),
				passingAssessmentPtr(),
			},
		},
	},
}

// TestEvaluate runs a series of tests on the ControlEvaluation.Evaluate method
func TestEvaluate(t *testing.T) {
	for _, test := range controlEvaluationTestData {
		t.Run(test.testName, func(t *testing.T) {
			c := test.control // copy the control to avoid duplication in the next test
			c.Evaluate(nil, testingApplicability, true)

			if c.Result.Status != test.expectedResult {
				t.Errorf("Expected Status to be %v, but it was %v", test.expectedResult, c.Result)
			}

			if c.CorruptedState != test.expectedCorrupted {
				t.Errorf("Expected CorruptedState to be %v, but it was %v", test.expectedCorrupted, c.CorruptedState)
			}
		})
		t.Run(test.testName+"no-changes", func(t *testing.T) {
			c := test.control // copy the control to avoid duplication in the next test
			c.Evaluate(nil, testingApplicability, false)

			for _, assessment := range c.Assessments {
				if assessment.Changes != nil {
					for _, change := range assessment.Changes {
						if change.Applied {
							t.Errorf("Expected no changes to be applied, but they were")
							return
						}
					}
				}
			}

			if c.Result.Status != test.expectedResult {
				t.Errorf("Expected Status to be %v, but it was %v", test.expectedResult, c.Result)
			}

			if c.CorruptedState != test.expectedCorrupted {
				t.Errorf("Expected CorruptedState to be %v, but it was %v", test.expectedCorrupted, c.CorruptedState)
			}
		})
	}
}

func TestAddAssessment(t *testing.T) {

	controlEvaluationTestData[0].control.AddAssessment("test", "test", []string{}, []AssessmentStep{})

	if controlEvaluationTestData[0].control.Result.Status != Failed {
		t.Errorf("Expected Status to be Failed, but it was %v", controlEvaluationTestData[0].control.Result)
	}

	if controlEvaluationTestData[0].control.Result.Message != "expected all AssessmentLog fields to have a value, but got: requirementId=len(4), description=len=(4), applicability=len(0), steps=len(0)" {
		t.Errorf("Expected error message to be 'expected all AssessmentLog fields to have a value, but got: requirementId=len(4), description=len=(4), applicability=len(0), steps=len(0)', but instead it was '%v'", controlEvaluationTestData[0].control.Result.Message)
	}

}
