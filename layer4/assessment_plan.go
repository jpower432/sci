package layer4

type AssessmentPlan struct {
	ControlId string `json:"control-id" yaml:"control-id"`

	Assessments []Assessment `json:"assessments" yaml:"assessments"`
}

type Assessment struct {
	RequirementId string `json:"requirement-id" yaml:"requirement-id"`

	Procedures []AssessmentProcedure `json:"procedures" yaml:"procedures"`
}

// AssessmentProcedure describes a testing procedure for evaluating a Layer 2 control requirement.
type AssessmentProcedure struct {
	// Id uniquely identifies the assessment procedure being executed
	Id string `json:"id" yaml:"id"`

	// Name provides a summary of the procedure
	Name string `json:"name" yaml:"name"`

	// Description provides a detailed explanation of the procedure
	Description string `json:"description" yaml:"description"`

	// Documentation provides a URL to documentation that describes how the assessment procedure evaluates the control requirement
	Documentation string `json:"documentation,omitempty" yaml:"documentation,omitempty"`
}

type Contact struct {
	// The contact person's name.
	Name string `json:"name" yaml:"name"`

	// Indicates whether this admin is the first point of contact for inquiries. Only one entry should be marked as primary.
	Primary bool `json:"primary" yaml:"primary"`

	// The entity with which the contact is affiliated, such as a school or employer.
	Affiliation *string `json:"affiliation,omitempty" yaml:"affiliation,omitempty"`

	// A preferred email address to reach the contact.
	Email *Email `json:"email,omitempty" yaml:"email,omitempty"`

	// A social media handle or profile for the contact.
	Social *string `json:"social,omitempty" yaml:"social,omitempty"`
}

type Email string
