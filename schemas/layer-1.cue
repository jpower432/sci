package schemas

@go(layer1)

#Catalog: {
    metadata?: #Metadata
    "framework-sections"?: [...#FrameworkSection] @go(FrameworkSections) @yaml("framework-sections,omitempty")
}

#Metadata: {
    id: string
    title: string
    description: string
    version?: string
    "last-modified"?: string @go(LastModified) @yaml("last-modified,omitempty")
    "publication-date"?: string @go(PublicationDate)
    "issuing-body"?: string @go(IssuingBody)
    "mapping-references"?: [...#MappingReference] @go(MappingReferences) @yaml("mapping-references,omitempty")
    resources?: [...#ResourceReference] @go(Resources)
}

#FrameworkSection: {
    title: string
    description: string
    guidelines?: [...#Guideline]
}

#Guideline: {
    id: string
    title?: string
    "guidance-text": string @go(GuidanceText)  @yaml("guidance-text")
    "guidance-sources"?: [...#Resource] @go(GuidanceSource )  @yaml("guidance-resources,omitempty")
    "guidelines-parts"?: [...#Guideline] @go(GuidelineParts) @yaml("guidance-parts,omitempty")
    // Control mapping to external catalogs
    "guideline-mapping"?: [...#Mapping] @go(GuidelineMappings)  @yaml("guideline-mappings,omitempty")
}

#MappingReference: {
    id: string
    title: string
    version: string
    description?: string
    url?: =~"^https?://[^\\s]+$"
}

#Mapping: {
    "reference-id": string @go(ReferenceId)
    identifiers: [...string]
    // Adding context about this particular mapping and why it was mapped.
    remarks?: string
}

#DocumentType: "Standard" | "Regulation" | "Best Practice"

#ResourceReference: {
	  id: string
    title: string
    description: string
    url?: =~"^https?://[^\\s]+$"
    "document-type"?: #DocumentType @go(DocumentType)
    "issuing-body"?: string @go(IssuingBody)
    "publication-date"?: string @go(PublicationDate)
}

#Resource: {
    "reference-id": string @go(ReferenceId) @yaml("reference-id")
    remarks?: string
}

