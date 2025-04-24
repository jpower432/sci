package layer2

func (r *AssessmentRequirement) IsApplicable(desiredApplicability []string) bool {
	for _, app := range r.Applicability {
		for _, desired := range desiredApplicability {
			if app == desired {
				return true
			}
		}
	}
	return false
}
