package layer4

// GetExecutorsForProcedure returns all executors that can execute a given procedure,
// along with their confidence levels.
func (ep *EvaluationPlan) GetExecutorsForProcedure(procedureID string) []ExecutorMapping {
	proc := ep.FindProcedure(procedureID)
	if proc == nil {
		return nil
	}
	return proc.Executors
}

// GetProceduresForExecutor returns all procedures that can be executed by a specific executor.
// This is useful for discovering which procedures a tool can run.
func (ep *EvaluationPlan) GetProceduresForExecutor(executorID string) []AssessmentProcedure {
	var procedures []AssessmentProcedure

	for _, plan := range ep.Plans {
		for _, assessment := range plan.Assessments {
			for i := range assessment.Procedures {
				proc := assessment.Procedures[i]
				// Check if executor is in the executors list
				for _, mapping := range proc.Executors {
					if mapping.Id == executorID {
						procedures = append(procedures, proc)
						break
					}
				}
			}
		}
	}

	return procedures
}

// GetExecutor returns an AssessmentExecutor by ID, or nil if not found.
func (ep *EvaluationPlan) GetExecutor(executorID string) *AssessmentExecutor {
	for i := range ep.Executors {
		if ep.Executors[i].Id == executorID {
			return &ep.Executors[i]
		}
	}
	return nil
}

// FindProcedure finds an AssessmentProcedure by ID within the evaluation plan.
// It searches through all assessment plans and their assessments to find the procedure.
func (ep *EvaluationPlan) FindProcedure(procedureID string) *AssessmentProcedure {
	for _, plan := range ep.Plans {
		for _, assessment := range plan.Assessments {
			for i := range assessment.Procedures {
				if assessment.Procedures[i].Id == procedureID {
					return &assessment.Procedures[i]
				}
			}
		}
	}
	return nil
}
