// Schema lifecycle: experimental | stable | deprecated
@status("stable")

package gemara

// EntryType enumerates the atomic units within Gemara artifacts that can participate in mappings
#EntryType: "guideline" | "control" | "assessment-requirement" | "threat" | "vector" @go(-)

// MappingDocument captures the user's intent for how entries across artifacts relate to one another
#MappingDocument: {
	// title describes the purpose of this mapping document at a glance
	title: string

	// metadata provides detailed data about this document
	"metadata": #Metadata & {
		"mapping-references": [#MappingReference, ...#MappingReference]
	} @go(Metadata)

	// mappings is one or more atomic relationships between entries in referenced artifacts
	mappings: [#Mapping, ...#Mapping] @go(Mappings)

	// remarks is prose regarding this mapping document
	remarks?: string
}

// Mapping represents an atomic relationship between two entries in referenced artifacts
#Mapping: {
	// source identifies the entry being mapped from
	source: #EntryReference @go(Source)

	// target identifies the entry being mapped to
	target: #EntryReference @go(Target)

	// type describes the nature of the relationship in the author's terms
	type: string @go(Type)

	// strength is the author's estimate of the relationship's intensity; Range: 1-10
	strength?: int & >=1 & <=10

	// remarks is prose describing this mapping relationship
	remarks?: string
}

// EntryReference identifies a specific entry within a referenced artifact
#EntryReference: {
	// reference-id points to a MappingReference in the document's metadata
	"reference-id": string @go(ReferenceId)

	// entry-id identifies the specific entry in the referenced artifact
	"entry-id": string @go(EntryId)

	// entry-type identifies what kind of atomic unit this entry is; extensible via string for external vocabularies
	"entry-type": #EntryType | string @go(EntryType)
}

// MappingReference represents a reference to an external document with full metadata.
#MappingReference: {
	// id allows this entry to be referenced by other elements
	id: string

	// title describes the purpose of this mapping reference at a glance
	title: string

	// version is the version identifier of the artifact being mapped to
	version: string

	// description is prose regarding the artifact's purpose or content
	description?: string

	// url is the path where the artifact may be retrieved; preferrably responds with Gemara-compatible YAML/JSON
	url?: =~"^(https?|file)://[^\\s]+$"
}
