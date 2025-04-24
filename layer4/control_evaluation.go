package layer4

func (c *ControlEvaluation) AddAssessment(requirementID string, methods []AssessmentMethod) (assessment *Assessment) {
	assessment = &Assessment{
		RequirementID: requirementID,
		Methods:       methods,
	}
	c.Assessments = append(c.Assessments, *assessment)
	return assessment
}

// Evaluate runs each AssessmentMethod in the assessment, updating the relevant fields on the control evaluation.
// `targetData` is the data that the assessment will be run against.
// `userApplicability` is a slice of strings that determine when the assessment is applicable. TODO: filtering which controls get evaluated sounds a lot like a job for the layer 3 Policy. Alternatively, should this take a Catalog and a Policy as arguments?
// `changesAllowed` determines whether the assessment is allowed to execute its changes.
func (c *ControlEvaluation) Evaluate(targetData interface{}, userApplicability []string, changesAllowed bool) {
	if len(c.Assessments) == 0 {
		return
	}

	// assessment requirements have applicability, a list of strings that define levels of adherence to the control catalog
	// unfortunately we don't have an AssessmentRequirement, we just have a reference to the requirementID value, with no way to load the catalog and source the required layer 2 data. I think this points even more to the need for this to be handled by layer 4 and 3 working together.
	// userApplicability is the set of those levels that the user selected for the evaluation
	// TODO
	// c.closeHandler()
	// for _, assessment := range c.Assessments {
	// 	var applicable bool
	// 	for _, aa := range assessment.Applicability {
	// 		for _, ua := range userApplicability {
	// 			if aa == ua {
	// 				applicable = true
	// 				break
	// 			}
	// 		}
	// 	}
	// 	if applicable {
	// 		result := assessment.Run(targetData, changesAllowed)
	// 		c.Result = UpdateAggregateResult(c.Result, result)
	// 		c.Message = assessment.Message
	// 		if c.Result == Failed {
	// 			break
	// 		}
	// 	}
	// }
	// TODO
	// c.Cleanup()
}

// func (c *ControlEvaluation) Cleanup() {
// 	for _, assessment := range c.Assessments {
// 		corrupted := assessment.RevertChanges()
// 		if corrupted {
// 			c.Corrupted_State = true
// 		}
// 	}
// }

// // CloseHandler creates a 'listener' on a new goroutine which will notify the
// // program if it receives an interrupt from the operating system.
// // If an interrupt is received, this will attempt to revert any changes
// // made by the terminated ControlEvaluation.
// func (c *ControlEvaluation) closeHandler() {
// 	// Ref: https://golangcode.com/handle-ctrl-c-exit-in-terminal/
// 	channel := make(chan os.Signal, 1)
// 	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
// 	go func() {
// 		<-channel
// 		log.Print("\n*****\nUnexpected termination. Attempting to revert changes made by the active ControlEvaluation. Do not interrupt this process.\n*****\n")
// 		c.Cleanup()
// 		os.Exit(0)
// 	}()
// }
