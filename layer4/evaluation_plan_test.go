package layer4

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExecutor(t *testing.T) {
	plan := &EvaluationPlan{
		Executors: []AssessmentExecutor{
			{Id: "scanner-a", Name: "Example Scanner A", Version: "1.0.0"},
			{Id: "scanner-b", Name: "Example Scanner B", Version: "2.1.0"},
		},
	}

	executor := plan.GetExecutor("scanner-a")
	require.NotNil(t, executor)
	assert.Equal(t, "Example Scanner A", executor.Name)

	executor = plan.GetExecutor("nonexistent")
	assert.Nil(t, executor)
}

func TestFindProcedure(t *testing.T) {
	plan := &EvaluationPlan{}
	plan.Plans = []struct {
		Control     Mapping      `json:"control"`
		Assessments []Assessment `json:"assessments"`
	}{
		{
			Assessments: []Assessment{
				{
					Procedures: []AssessmentProcedure{
						{Id: "proc-1", Name: "Procedure 1"},
						{Id: "proc-2", Name: "Procedure 2"},
					},
				},
			},
		},
	}

	proc := plan.FindProcedure("proc-1")
	require.NotNil(t, proc)
	assert.Equal(t, "Procedure 1", proc.Name)

	proc = plan.FindProcedure("nonexistent")
	assert.Nil(t, proc)
}

func TestGetExecutorsForProcedure(t *testing.T) {
	plan := &EvaluationPlan{
		Executors: []AssessmentExecutor{
			{Id: "scanner-a", Name: "Example Scanner A"},
			{Id: "scanner-b", Name: "Example Scanner B"},
		},
	}
	plan.Plans = []struct {
		Control     Mapping      `json:"control"`
		Assessments []Assessment `json:"assessments"`
	}{
		{
			Assessments: []Assessment{
				{
					Procedures: []AssessmentProcedure{
						{
							Id: "scan-image",
							Executors: []ExecutorMapping{
								{Id: "scanner-a", TrustScore: 8},
								{Id: "scanner-b", TrustScore: 7},
							},
						},
					},
				},
			},
		},
	}

	executors := plan.GetExecutorsForProcedure("scan-image")
	assert.Len(t, executors, 2)
	executor := executors[0]

	assert.Equal(t, "scanner-a", executor.Id)
	assert.Equal(t, int64(8), executor.TrustScore)
}
