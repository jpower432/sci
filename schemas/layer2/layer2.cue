package layer2

import "github.com/ossf/gemara/schemas/common"

@go(layer2)

#ControlObjectives: {
	metadata?: #Metadata

	families: [...common.#Family] @go(Families)
	controls?: [...#Control] @go(Controls)
	threats?: [...#Threat] @go(Threats)
	capabilities?: [...#Capability] @go(Capabilities)

	"imported-controls"?: [...#Mapping] @go(ImportedControls)
	"imported-threats"?: [...#Mapping] @go(ImportedThreats)
	"imported-capabilities"?: [...#Mapping] @go(ImportedCapabilities)
}

// Resuable types //
#Metadata: {
	id:               string
	title:            string
	description:      string
	version?:         string
	"last-modified"?: string @go(LastModified) @yaml("last-modified,omitempty")
	"applicability-categories"?: [...#Category] @go(ApplicabilityCategories) @yaml("applicability-categories,omitempty")
	"mapping-references"?: [...#MappingReference] @go(MappingReferences) @yaml("mapping-references,omitempty")
}

#Category: {
	id:          string
	title:       string
	description: string
}

#Control: {
	id:           string
	title:        string
	objective:    string
	"family-id"?: string @go(FamilyId) @yaml("family-id,omitempty")
	"assessment-requirements": [...#AssessmentRequirement] @go(AssessmentRequirements)
	"guideline-mappings"?: [...#Mapping] @go(GuidelineMappings)
	"threat-mappings"?: [...#Mapping] @go(ThreatMappings)
}

#Threat: {
	id:          string
	title:       string
	description: string
	capabilities: [...#Mapping]

	"external-mappings"?: [...#Mapping] @go(ExternalMappings)
}

#Capability: {
	id:          string
	title:       string
	description: string
}

// MappingReference uses the common definition
#MappingReference: common.#MappingReference

// Mapping uses the common definition
#Mapping: common.#Mapping

// MappingEntry uses the common definition
#MappingEntry: common.#MappingEntry

#AssessmentRequirement: {
	id:   string
	text: string
	applicability: [...string]

	recommendation?: string
}
