package layer3

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ossf/gemara/layer2"
)

// ChecklistItem represents a single checklist item.
type ChecklistItem struct {
	// RequirementId is the requirement ID.
	RequirementId string
	// Requirement is the human-readable requirement.
	Requirement string
	// Description provides additional context or a summary about the requirement.
	Description string
	// Documentation is the documentation URL
	Documentation string
}

// ControlSection organizes checklist items by control.
type ControlSection struct {
	// ControlName is the control identifier (e.g., "AC-1")
	ControlName string
	// ControlReference is the formatted reference (e.g., "NIST-800-53 / AC-1")
	ControlReference string
	// Items are the checklist items for this control
	Items []ChecklistItem
}

// Checklist represents the structured checklist data.
type Checklist struct {
	// PolicyId identifies the policy (using metadata ID).
	PolicyId string
	// Author is the name of the policy author.
	Author string
	// AuthorVersion is the version of the authoring tool or system.
	AuthorVersion string
	// Sections are the applicability sections
	Sections []ControlSection
}

// ToChecklist converts a Policy into a structured Checklist.
func (p *Policy) ToChecklist() (Checklist, error) {
	checklist := Checklist{
		PolicyId:      p.Metadata.Id,
		Author:        p.Metadata.Author.Name,
		AuthorVersion: p.Metadata.Author.Version,
	}

	processPolicyMapping := func(mapping PolicyMapping) error {
		if mapping.ReferenceId == "" {
			return nil
		}

		sections, err := buildChecklistItems(p, &mapping)
		if err != nil {
			return fmt.Errorf("failed to build checklist items for reference %q: %w", mapping.ReferenceId, err)
		}

		checklist.Sections = append(checklist.Sections, sections...)
		return nil
	}

	for _, controlRef := range p.ControlReferences {
		if err := processPolicyMapping(controlRef); err != nil {
			return Checklist{}, err
		}
	}

	return checklist, nil
}

// ToMarkdownChecklist converts a Policy into a markdown checklist.
// Generates a pre-execution checklist showing what needs to be checked.
func (p *Policy) ToMarkdownChecklist() (string, error) {
	checklist, err := p.ToChecklist()
	if err != nil {
		return "", fmt.Errorf("failed to build checklist: %w", err)
	}

	tmpl, err := template.New("checklist").Parse(markdownTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, checklist); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// buildChecklistItems converts a PolicyMapping into checklist items from Layer 2 Controls.
// It loads the Layer 2 catalog from the reference-id, finds each control, applies modifications,
// and creates checklist items from the assessment requirements.
func buildChecklistItems(policy *Policy, mapping *PolicyMapping) ([]ControlSection, error) {
	if mapping == nil {
		return nil, fmt.Errorf("policy mapping is nil")
	}
	if policy == nil {
		return nil, fmt.Errorf("policy is nil")
	}

	var catalogSource string
	for _, ref := range policy.Metadata.MappingReferences {
		if ref.Id == mapping.ReferenceId {
			catalogSource = ref.Url
			break
		}
	}

	if catalogSource == "" {
		return nil, fmt.Errorf("mapping reference %q not found in policy metadata", mapping.ReferenceId)
	}

	catalog := &layer2.ControlObjectives{}
	err := catalog.LoadFile(catalogSource)
	if err != nil {
		return nil, fmt.Errorf("failed to load catalog for reference %q from %qL: %w)", mapping.ReferenceId, catalogSource, err)
	}

	var sections []ControlSection
	for _, controlMod := range mapping.ControlModifications {
		if controlMod.TargetId == "" {
			continue
		}
		
		var baseControl *layer2.Control
		for i := range catalog.Controls {
			if catalog.Controls[i].Id == controlMod.TargetId {
				baseControl = &catalog.Controls[i]
				break
			}
		}

		if baseControl == nil {
			return nil, fmt.Errorf("control %q not found in catalog %q", controlMod.TargetId, mapping.ReferenceId)
		}

		// Apply modifications to create the final control
		finalControl := *baseControl
		if controlMod.Overrides != nil {
			// Apply overrides
			if controlMod.Overrides.Title != "" {
				finalControl.Title = controlMod.Overrides.Title
			}
			if controlMod.Overrides.Objective != "" {
				finalControl.Objective = controlMod.Overrides.Objective
			}
			if len(controlMod.Overrides.AssessmentRequirements) > 0 {
				finalControl.AssessmentRequirements = controlMod.Overrides.AssessmentRequirements
			}
		}

		// Create a new control section
		section := ControlSection{
			ControlName:      controlMod.TargetId,
			ControlReference: fmt.Sprintf("%s / %s", mapping.ReferenceId, controlMod.TargetId),
			Items:            []ChecklistItem{},
		}

		// Create checklist items from assessment requirements
		for _, req := range finalControl.AssessmentRequirements {
			item := ChecklistItem{
				RequirementId: req.Id,
				Requirement:   req.Text,
				Description:   req.Text,
			}

			// Add recommendation if available
			if req.Recommendation != "" {
				if item.Description != "" {
					item.Description = fmt.Sprintf("%s Recommendation: %s", item.Description, req.Recommendation)
				} else {
					item.Description = fmt.Sprintf("Recommendation: %s", req.Recommendation)
				}
			}

			section.Items = append(section.Items, item)
		}

		// Apply assessment requirement modifications nested under this control
		for _, reqMod := range controlMod.AssessmentRequirementModifications {
			if reqMod.TargetId == "" {
				continue
			}

			// Find the matching checklist item in this control section
			for itemIdx := range section.Items {
				// Match by exact requirement ID
				itemReqId := section.Items[itemIdx].RequirementId
				fullReqId := fmt.Sprintf("%s.%s", controlMod.TargetId, itemReqId)

				if itemReqId == reqMod.TargetId || fullReqId == reqMod.TargetId {
					// Update the checklist item based on the modification
					if reqMod.Overrides != nil {
						if reqMod.Overrides.Text != "" {
							section.Items[itemIdx].Requirement = reqMod.Overrides.Text
							section.Items[itemIdx].Description = reqMod.Overrides.Text
						}
						if reqMod.Overrides.Recommendation != "" {
							if section.Items[itemIdx].Description != "" {
								section.Items[itemIdx].Description = fmt.Sprintf("%s Recommendation: %s", section.Items[itemIdx].Description, reqMod.Overrides.Recommendation)
							} else {
								section.Items[itemIdx].Description = fmt.Sprintf("Recommendation: %s", reqMod.Overrides.Recommendation)
							}
						}
					}

					if reqMod.ModificationRationale != "" {
						if section.Items[itemIdx].Description != "" {
							section.Items[itemIdx].Description = fmt.Sprintf("%s. %s", reqMod.ModificationRationale, section.Items[itemIdx].Description)
						} else {
							section.Items[itemIdx].Description = reqMod.ModificationRationale
						}
					}

					break
				}
			}
		}

		// Only add section if it has items
		if len(section.Items) > 0 {
			sections = append(sections, section)
		}
	}

	return sections, nil
}
