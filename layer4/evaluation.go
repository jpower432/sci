package layer4

import "github.com/revanite-io/sci/layer2"

// The Evaluation type is defined in `generated_types.go`

// TODO: add a `policy layer3.Policy` argument. If we had a policy here, then the buildControlEvaluations
// could apply policy to each control

// NewEvaluation creates a new Evaluation instance based on the provided catalog.
// It initializes the Evaluation with the framework ID from the catalog and builds control
// evaluations for every control in the catalog. Each ControlEvaluation is initialized with
// the control family name and control ID and one Assessment for each requirement in the control.
func NewEvaluation(catalog layer2.Catalog) *Evaluation {
	eval := &Evaluation{
		CatalogID: catalog.Metadata.Id,
	}
	eval.buildControlEvaluations(catalog)
	return eval
}

func (e *Evaluation) buildControlEvaluations(catalog layer2.Catalog) {
	for _, cf := range catalog.ControlFamilies {
		for _, c := range cf.Controls {
			eval := ControlEvaluation{
				ControlID: c.Id,
			}

			for _, r := range c.AssessmentRequirements {
				// TODO: add a `policy layer3.Policy` argument to this func. If we had a policy here, then we could do
				// something like:
				// if Policy.AssessmentRequired(requirement) {
				// 	to not add assessments when the policy says they are not relevant or required
				eval.Assessments = append(eval.Assessments, Assessment{
					RequirementID: r.Id,
				})
			}
			e.ControlEvaluations = append(e.ControlEvaluations, eval)
		}
	}
}
