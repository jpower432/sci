package layer1

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
)

const (
	oscalVersion    = "1.1.3"
	gemaraNamespace = "https://github.com/ossf/gemara/ns/oscal"
)

// ToOSCALProfile creates an OSCAL Profile from the shared guidelines in a given
// Layer 1 Guidance Document.
func (g *GuidanceDocument) ToOSCALProfile() (oscalTypes.Profile, error) {
	metadata, err := createMetadata(g)
	if err != nil {
		return oscalTypes.Profile{}, fmt.Errorf("error creating profile metadata: %w", err)
	}

	mappingSet := make(map[string]oscalTypes.Import)
	for _, mappingRef := range g.Metadata.MappingReferences {
		mappingSet[mappingRef.Id] = oscalTypes.Import{
			Href: mappingRef.Url,
		}
	}

	for _, mapping := range g.ImportedGuidelines {
		targetImport, ok := mappingSet[mapping.ReferenceId]
		if !ok {
			continue
		}

		withIds := make([]string, 0, len(mapping.Entries))
		for _, entry := range mapping.Entries {
			withIds = append(withIds, normalizeControl(entry.ReferenceId))
		}

		selector := oscalTypes.SelectControlById{
			WithIds: &withIds,
		}
		targetImport.IncludeControls = &[]oscalTypes.SelectControlById{selector}
		mappingSet[mapping.ReferenceId] = targetImport
	}

	imports := make([]oscalTypes.Import, 0, len(mappingSet))
	for _, imp := range mappingSet {
		if imp.IncludeControls == nil {
			continue
		}
		imports = append(imports, imp)
	}

	if len(imports) == 0 {
		return oscalTypes.Profile{}, fmt.Errorf("guidance %s: imported_guidelines cannot be empty", g.Metadata.Id)
	}

	profile := oscalTypes.Profile{
		UUID:     uuid.NewUUID(),
		Imports:  imports,
		Metadata: metadata,
	}
	return profile, nil
}

// ToOSCALCatalog creates an OSCAL Catalog from the locally defined guidelines in a given
// Layer 1 Guidance Document.
func (g *GuidanceDocument) ToOSCALCatalog() (oscalTypes.Catalog, error) {
	metadata, err := createMetadata(g)
	if err != nil {
		return oscalTypes.Catalog{}, fmt.Errorf("error creating catalog metadata: %w", err)
	}

	// Create a resource map for control linking
	resourcesMap := make(map[string]string)
	backmatter := resourcesToBackMatter(g.Metadata.Resources)
	if backmatter != nil {
		for _, resource := range *backmatter.Resources {
			// Extract the id from the props
			props := *resource.Props
			id := props[0].Value
			resourcesMap[id] = resource.UUID
		}
	}

	var groups []oscalTypes.Group
	for _, category := range g.Categories {
		groups = append(groups, g.createControlGroup(category, resourcesMap))
	}

	catalog := oscalTypes.Catalog{
		UUID:       uuid.NewUUID(),
		Metadata:   metadata,
		Groups:     nilIfEmpty(&groups),
		BackMatter: backmatter,
	}
	return catalog, nil
}

func createMetadata(guidance *GuidanceDocument) (oscalTypes.Metadata, error) {
	metadata := oscalTypes.Metadata{
		Title:        guidance.Metadata.Title,
		OscalVersion: oscalVersion,
		Version:      guidance.Metadata.Version,
	}

	if guidance.Metadata.PublicationDate != "" {
		published, err := time.Parse(time.DateOnly, guidance.Metadata.PublicationDate)
		if err != nil {
			return oscalTypes.Metadata{}, err
		}
		metadata.Published = &published
	}

	var err error
	lastModified := time.Now()
	if guidance.Metadata.LastModified != "" {
		lastModified, err = time.Parse(time.DateTime, guidance.Metadata.LastModified)
		if err != nil {
			return oscalTypes.Metadata{}, err
		}
	}
	metadata.LastModified = lastModified

	authorRole := oscalTypes.Role{
		ID:          "author",
		Description: "Author of the guidance document",
		Title:       "Author",
	}

	author := oscalTypes.Party{
		UUID: uuid.NewUUID(),
		Type: "person",
		Name: guidance.Metadata.Author,
	}

	responsibleParty := oscalTypes.ResponsibleParty{
		PartyUuids: []string{author.UUID},
		RoleId:     authorRole.ID,
	}

	metadata.Parties = &[]oscalTypes.Party{author}
	metadata.Roles = &[]oscalTypes.Role{authorRole}
	metadata.ResponsibleParties = &[]oscalTypes.ResponsibleParty{responsibleParty}
	return metadata, nil
}

func (g *GuidanceDocument) createControlGroup(category Category, resourcesMap map[string]string) oscalTypes.Group {
	group := oscalTypes.Group{
		Class: "category",
		ID:    category.Id,
		Title: category.Title,
	}

	controlMap := make(map[string]oscalTypes.Control)
	for _, guideline := range category.Guidelines {
		control, parent := g.guidelineToControl(guideline, resourcesMap)

		if parent == "" {
			controlMap[control.ID] = control
		} else {
			parentControl := controlMap[parent]
			if parentControl.Controls == nil {
				parentControl.Controls = &[]oscalTypes.Control{}
			}
			*parentControl.Controls = append(*parentControl.Controls, control)
			controlMap[parent] = parentControl
		}
	}

	controls := make([]oscalTypes.Control, 0, len(controlMap))
	for _, control := range controlMap {
		controls = append(controls, control)
	}

	group.Controls = nilIfEmpty(&controls)
	return group
}

func (g *GuidanceDocument) guidelineToControl(guideline Guideline, resourcesMap map[string]string) (oscalTypes.Control, string) {
	controlId := normalizeControl(guideline.Id)

	control := oscalTypes.Control{
		ID:    controlId,
		Title: guideline.Title,
		Class: g.Metadata.Id,
	}

	var links []oscalTypes.Link
	for _, also := range guideline.SeeAlso {
		relatedLink := oscalTypes.Link{
			Href: fmt.Sprintf("#%s", normalizeControl(also)),
			Rel:  "related",
		}
		links = append(links, relatedLink)
	}

	for _, external := range guideline.ExternalReferences {
		ref, found := resourcesMap[external]
		if !found {
			continue
		}
		externalLink := oscalTypes.Link{
			Href: fmt.Sprintf("#%s", ref),
			Rel:  "reference",
		}
		links = append(links, externalLink)
	}
	control.Links = nilIfEmpty(&links)

	// Top-level statements are required for controls per OSCAL guidance
	smtPart := oscalTypes.Part{
		Name: "statement",
		ID:   fmt.Sprintf("%s_smt", controlId),
	}
	var subSmts []oscalTypes.Part
	for _, part := range guideline.GuidelineParts {

		partId := normalizeControl(part.Id)
		lastDotIndex := strings.LastIndex(partId, ".")
		if lastDotIndex != -1 && lastDotIndex < len(partId)-1 {
			partId = partId[lastDotIndex+1:]
		}

		subSmt := oscalTypes.Part{
			Name:  "item",
			ID:    fmt.Sprintf("%s_smt.%s", controlId, partId),
			Prose: part.Prose,
			Title: part.Title,
		}

		if len(part.Recommendations) > 0 {
			gdnSubPart := oscalTypes.Part{
				Name:  "guidance",
				ID:    fmt.Sprintf("%s_smt.%s_gdn", controlId, partId),
				Prose: strings.Join(part.Recommendations, " "),
			}
			subSmt.Parts = &[]oscalTypes.Part{
				gdnSubPart,
			}
		}

		subSmts = append(subSmts, subSmt)
	}
	smtPart.Parts = nilIfEmpty(&subSmts)
	control.Parts = &[]oscalTypes.Part{smtPart}

	if guideline.Objective != "" {
		// objective part
		objPart := oscalTypes.Part{
			Name:  "assessment-objective",
			ID:    fmt.Sprintf("%s_obj", controlId),
			Prose: guideline.Objective,
		}
		*control.Parts = append(*control.Parts, objPart)
	}

	if len(guideline.Recommendations) > 0 {
		// gdn part
		gdnPart := oscalTypes.Part{
			Name:  "guidance",
			ID:    fmt.Sprintf("%s_gdn", controlId),
			Prose: strings.Join(guideline.Recommendations, " "),
		}
		*control.Parts = append(*control.Parts, gdnPart)
	}

	return control, normalizeControl(guideline.BaseGuidelineID)
}

func resourcesToBackMatter(resourceRefs []ResourceReference) *oscalTypes.BackMatter {
	var resources []oscalTypes.Resource
	for _, ref := range resourceRefs {
		resource := oscalTypes.Resource{
			UUID:        uuid.NewUUID(),
			Title:       ref.Title,
			Description: ref.Description,
			Props: &[]oscalTypes.Property{
				{
					Name:  "id",
					Value: ref.Id,
					Ns:    gemaraNamespace,
				},
			},
			Rlinks: &[]oscalTypes.ResourceLink{
				{
					Href: ref.Url,
				},
			},
			Citation: &oscalTypes.Citation{
				Text: fmt.Sprintf(
					"%s. (%s). *%s*. %s",
					ref.IssuingBody,
					ref.PublicationDate,
					ref.Title,
					ref.Url),
			},
		}
		resources = append(resources, resource)
	}

	if len(resources) == 0 {
		return nil
	}

	backmatter := oscalTypes.BackMatter{
		Resources: &resources,
	}
	return &backmatter
}

func nilIfEmpty[T any](slice *[]T) *[]T {
	if slice == nil || len(*slice) == 0 {
		return nil
	}
	return slice
}

func normalizeControl(input string) string {
	re := regexp.MustCompile(`\((\d+)\)`)
	replacedString := re.ReplaceAllString(input, ".$1")
	finalString := strings.ToLower(replacedString)
	return finalString
}
