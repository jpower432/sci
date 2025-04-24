package layer2

import "testing"

func TestIsApplicable(t *testing.T) {
	tests := []struct {
		name               string
		requirement        AssessmentRequirement
		applicability      []string
		expectedApplicable bool
	}{
		{
			name: "Single match",
			requirement: AssessmentRequirement{
				Id:            "1",
				Text:          "Test Requirement",
				Applicability: []string{"applicable"},
			},
			applicability:      []string{"applicable"},
			expectedApplicable: true,
		},
		{
			name: "No match",
			requirement: AssessmentRequirement{
				Id:            "2",
				Text:          "Test Requirement 2",
				Applicability: []string{"not_applicable"},
			},
			applicability:      []string{"applicable"},
			expectedApplicable: false,
		},
		{
			name: "Multiple applicability",
			requirement: AssessmentRequirement{
				Id:            "3",
				Text:          "Test Requirement 3",
				Applicability: []string{"applicable", "extra_applicable", "incredibly_applicable"},
			},
			applicability:      []string{"extra_applicable"},
			expectedApplicable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.requirement.IsApplicable(tt.applicability)
			if result != tt.expectedApplicable {
				t.Errorf("IsApplicable() = %v, want %v", result, tt.expectedApplicable)
			}
		})
	}
}
