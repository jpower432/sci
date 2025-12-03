package layer2

import (
	"testing"

	oscal "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/stretchr/testify/assert"

	"github.com/ossf/gemara/common"
	oscalUtils "github.com/ossf/gemara/internal/oscal"
)

var TestCases = []struct {
	name          string
	catalog       *ControlObjectives
	controlHREF   string
	wantErr       bool
	expectedTitle string
}{
	{
		name: "Valid catalog with single control family",
		catalog: &ControlObjectives{
			Metadata: Metadata{
				Id:      "test-catalog",
				Title:   "Test Catalog",
				Version: "devel",
			},
			Families: []common.Family{
				{
					Id:          "AC",
					Title:       "access-control",
					Description: "Controls for access management",
				},
			},
			Controls: []Control{
				{
					Id:       "AC-01",
					Title:    "Access Control Policy",
					FamilyId: "AC",
					AssessmentRequirements: []AssessmentRequirement{
						{
							Id:   "AC-01.1",
							Text: "Develop and document access control policy",
						},
					},
				},
			},
		},
		controlHREF:   "https://baseline.openssf.org/versions/%s#%s",
		wantErr:       false,
		expectedTitle: "Test Catalog",
	},
	{
		name: "Valid catalog with multiple control families",
		catalog: &ControlObjectives{
			Metadata: Metadata{
				Id:      "test-catalog-multi",
				Title:   "Test Catalog Multiple",
				Version: "devel",
			},
			Families: []common.Family{
				{
					Id:          "AC",
					Title:       "access-control",
					Description: "Controls for access management",
				},
				{
					Id:          "BR",
					Title:       "business-requirements",
					Description: "Controls for business requirements",
				},
			},
			Controls: []Control{
				{
					Id:       "AC-01",
					Title:    "Access Control Policy",
					FamilyId: "AC",
					AssessmentRequirements: []AssessmentRequirement{
						{
							Id:   "AC-01.1",
							Text: "Develop and document access control policy",
						},
					},
				},
				{
					Id:       "BR-01",
					Title:    "Business Requirements Policy",
					FamilyId: "BR",
					AssessmentRequirements: []AssessmentRequirement{
						{
							Id:   "BR-01.1",
							Text: "Define business requirements",
						},
					},
				},
			},
		},
		controlHREF:   "https://baseline.openssf.org/versions/%s#%s",
		wantErr:       false,
		expectedTitle: "Test Catalog Multiple",
	},
	{
		name: "Valid catalog with controls without a family",
		catalog: &ControlObjectives{
			Metadata: Metadata{
				Id:      "test-catalog-no-family",
				Title:   "Test Catalog No Family",
				Version: "devel",
			},
			Families: []common.Family{
				{
					Id:          "AC",
					Title:       "access-control",
					Description: "Controls for access management",
				},
			},
			Controls: []Control{
				{
					Id:       "AC-01",
					Title:    "Access Control Policy",
					FamilyId: "AC",
					Objective: "Ensure access control",
					AssessmentRequirements: []AssessmentRequirement{
						{
							Id:   "AC-01.1",
							Text: "Develop and document access control policy",
						},
					},
				},
				{
					Id:        "ORPHAN-01",
					Title:     "Orphan Control",
					FamilyId:  "", // No family
					Objective: "Standalone control",
					AssessmentRequirements: []AssessmentRequirement{
						{
							Id:   "ORPHAN-01.1",
							Text: "Orphan requirement",
						},
					},
				},
			},
		},
		controlHREF:   "https://baseline.openssf.org/versions/%s#%s",
		wantErr:       false,
		expectedTitle: "Test Catalog No Family",
	},
}

func Test_toOSCAL(t *testing.T) {
	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			oscalCatalog, err := tt.catalog.ToOSCAL(tt.controlHREF)

			if (err == nil) == tt.wantErr {
				t.Errorf("ToOSCAL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Wrap oscal catalog
			// Create the proper OSCAL document structure
			oscalDocument := oscal.OscalModels{
				Catalog: &oscalCatalog,
			}

			// Create validation for the OSCAL catalog
			assert.NoError(t, oscalUtils.Validate(oscalDocument))

			// Compare each field
			assert.NotEmpty(t, oscalCatalog.UUID)
			assert.Equal(t, tt.expectedTitle, oscalCatalog.Metadata.Title)
			assert.Equal(t, tt.catalog.Metadata.Version, oscalCatalog.Metadata.Version)
			
			// Count unique families (only those that exist in Families array)
			familyMap := make(map[string]bool)
			familyById := make(map[string]bool)
			for _, family := range tt.catalog.Families {
				familyById[family.Id] = true
			}
			
			controlsWithFamily := 0
			controlsWithoutFamily := 0
			for _, control := range tt.catalog.Controls {
				if control.FamilyId != "" && familyById[control.FamilyId] {
					familyMap[control.FamilyId] = true
					controlsWithFamily++
				} else {
					controlsWithoutFamily++
				}
			}
			
			// Groups should match families that exist
			if oscalCatalog.Groups != nil {
				assert.Equal(t, len(familyMap), len(*oscalCatalog.Groups))
			} else {
				assert.Equal(t, 0, len(familyMap), "No groups expected when all controls have families")
			}
			
			// Controls without families should be in catalog.Controls
			if controlsWithoutFamily > 0 {
				assert.NotNil(t, oscalCatalog.Controls, "Controls without families should be in catalog.Controls")
				if oscalCatalog.Controls != nil {
					assert.Equal(t, controlsWithoutFamily, len(*oscalCatalog.Controls))
				}
			}

			// Compare each control family group
			groups := (*oscalCatalog.Groups)
			for i, group := range groups {
				// Verify group exists in our family map
				found := false
				for familyId := range familyMap {
					if group.ID == familyId {
						found = true
						break
					}
				}
				assert.True(t, found, "Group %d with ID %s should exist in catalog", i, group.ID)
			}
		})
	}
}
