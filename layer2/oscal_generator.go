package layer2

import (
	"fmt"
	"strings"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscal "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"

	"github.com/ossf/gemara/common"
	oscalUtils "github.com/ossf/gemara/internal/oscal"
)

const defaultVersion = "0.0.1"

// ToOSCAL converts a ControlObjectives to OSCAL Catalog format.
// Parameters:
//   - controlHREF: URL template for linking to controls. Uses format: controlHREF(version, controlID)
//     Example: "https://baseline.openssf.org/versions/%s#%s"
//
// The function automatically:
//   - Uses the control objectives' internal version from Metadata.Version
//   - Groups controls by Family.Id and uses it as the OSCAL group ID
//   - Generates a unique UUID for the catalog
func (c *ControlObjectives) ToOSCAL(controlHREF string) (oscal.Catalog, error) {
	now := time.Now()

	version := c.Metadata.Version
	if c.Metadata.Version == "" {
		version = defaultVersion
	}

	oscalCatalog := oscal.Catalog{
		UUID:     uuid.NewUUID(),
		Groups:   nil,
		Controls: nil,
		Metadata: oscal.Metadata{
			LastModified: oscalUtils.GetTimeWithFallback(c.Metadata.LastModified, now),
			Links: &[]oscal.Link{
				{
					Href: fmt.Sprintf(controlHREF, c.Metadata.Version, ""),
					Rel:  "canonical",
				},
			},
			OscalVersion: oscal.Version,
			Published:    &now,
			Title:        c.Metadata.Title,
			Version:      version,
		},
	}

	catalogGroups := []oscal.Group{}

	// Build a map of families by ID
	familyById := make(map[string]common.Family)
	for _, family := range c.Families {
		familyById[family.Id] = family
	}

	// Group controls by family ID
	familyMap := make(map[string][]Control)
	for _, control := range c.Controls {
		familyId := "none"
		if control.FamilyId != "" {
			familyId = control.FamilyId
		}
		familyMap[familyId] = append(familyMap[familyId], control)
	}

	var catalogControls []oscal.Control

	// Create OSCAL groups for each family and handle controls without a family
	for familyId, controls := range familyMap {
		if family, exists := familyById[familyId]; exists {
			group := oscal.Group{
				Class:    "family",
				Controls: nil,
				ID:       familyId,
				Title:    strings.ReplaceAll(family.Description, "\n", "\\n"),
			}

			oscalControls := []oscal.Control{}
			for _, control := range controls {
				oscalControl := c.controlToOSCAL(control, familyId, controlHREF)
				oscalControls = append(oscalControls, oscalControl)
			}

			group.Controls = &oscalControls
			catalogGroups = append(catalogGroups, group)
		} else {
			// Controls without a family are added directly to the catalog (not in a group)
			for _, control := range controls {
				oscalControl := c.controlToOSCAL(control, "", controlHREF)
				catalogControls = append(catalogControls, oscalControl)
			}
		}
	}

	oscalCatalog.Groups = oscalUtils.NilIfEmpty(catalogGroups)
	oscalCatalog.Controls = oscalUtils.NilIfEmpty(catalogControls)

	return oscalCatalog, nil
}

// controlToOSCAL converts a single Control to an OSCAL Control
func (c *ControlObjectives) controlToOSCAL(control Control, familyId string, controlHREF string) oscal.Control {
	controlTitle := strings.TrimSpace(control.Title)

	newCtl := oscal.Control{
		Class: familyId,
		ID:    control.Id,
		Title: strings.ReplaceAll(controlTitle, "\n", "\\n"),
		Parts: &[]oscal.Part{
			{
				Name:  "statement",
				ID:    fmt.Sprintf("%s_smt", control.Id),
				Prose: control.Objective,
			},
		},
		Links: &[]oscal.Link{
			{
				Href: fmt.Sprintf(controlHREF, c.Metadata.Version, strings.ToLower(control.Id)),
				Rel:  "canonical",
			},
		},
	}

	var subControls []oscal.Control
	for _, ar := range control.AssessmentRequirements {
		subControl := oscal.Control{
			ID:    ar.Id,
			Title: ar.Id,
			Parts: &[]oscal.Part{
				{
					Name:  "statement",
					ID:    fmt.Sprintf("%s_smt", ar.Id),
					Prose: ar.Text,
				},
			},
		}

		if ar.Recommendation != "" {
			*subControl.Parts = append(*subControl.Parts, oscal.Part{
				Name:  "guidance",
				ID:    fmt.Sprintf("%s_gdn", ar.Id),
				Prose: ar.Recommendation,
			})
		}

		*subControl.Parts = append(*subControl.Parts, oscal.Part{
			Name: "assessment-objective",
			ID:   fmt.Sprintf("%s_obj", ar.Id),
			Links: &[]oscal.Link{
				{
					Href: fmt.Sprintf("#%s_smt", ar.Id),
					Rel:  "assessment-for",
				},
			},
		})

		subControls = append(subControls, subControl)
	}

	if len(subControls) > 0 {
		newCtl.Controls = &subControls
	}

	return newCtl
}
