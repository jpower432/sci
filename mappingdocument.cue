// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// MappingDocument captures the user's intent for how entries in a source artifact relate to entries in a target artifact
#MappingDocument: {
	// title describes the purpose of this mapping document at a glance
	title: string

	// metadata provides detailed data about this document
	metadata: #Metadata @go(Metadata)
	metadata: type: "MappingDocument"
	metadata: "mapping-references": [#MappingReference, ...#MappingReference]

	// source-reference identifies the artifact being mapped from; must match a mapping-reference id
	"source-reference": #ArtifactMapping @go(SourceReference)

	// target-reference identifies the artifact being mapped to; must match a mapping-reference id
	"target-reference": #ArtifactMapping @go(TargetReference)

	// mappings is one or more atomic relationships between entries in the referenced artifacts
	mappings: [#_MappingStrict, ...#_MappingStrict] @go(Mappings,type=[]Mapping)

	_uniqueMappingIds: {for i, m in mappings {(m.id): i}}

	// remarks is prose regarding this mapping document
	remarks?: string
}

// _MappingStrict layers the "target required when not no-match" rule on top of #Mapping
#_MappingStrict: {
	@go(-)
} & #Mapping & {
	relationship: #RelationshipType
	if relationship != "no-match" {
		target: #TypedEntry
	}
}

// Mapping represents an atomic relationship between a source entry and an optional target entry
#Mapping: {
	// id allows this mapping to be referenced by other elements
	id: string

	// source identifies the entry being mapped from
	source: #TypedEntry @go(Source)

	// target identifies the entry being mapped to; absent when relationship is no-match
	target?: #TypedEntry @go(Target,optional=nillable)

	// relationship describes the nature or purpose of the mapping
	relationship: #RelationshipType @go(Relationship)

	// strength is the author's estimate of how completely the source entry satisfies the target entry; range 1-10
	strength?: int & >=1 & <=10 @go(Strength)

	"confidence-level"?: #ConfidenceLevel @go(ConfidenceLevel)

	// applicability constrains the contexts in which this mapping holds
	applicability?: [string, ...string] @go(Applicability)

	// rationale explains why this relationship exists
	rationale?: string

	// remarks is general prose regarding this mapping
	remarks?: string
}

// RelationshipType enumerates the nature of the mapping between entries.
#RelationshipType:
	// source fulfills the target's objective
	"implements" |
	// target fulfills the source's objective (requirements-to-implementation direction)
	"implemented-by" |
	// source contributes to, but does not fully satisfy, the target
	"supports" |
	// target contributes to, but does not fully satisfy, the source
	"supported-by" |
	// source and target express the same intent
	"equivalent" |
	// source fully contains the target's scope and more
	"subsumes" |
	// source has no counterpart in the target artifact
	"no-match" |
	// source and target are related but the nature is unspecified
	"relates-to" @go(-)

// EntryReference identifies a specific entry within a referenced artifact
#TypedEntry: {
	// entry-id identifies the specific entry in the referenced artifact
	"entry-id": string @go(EntryId)

	// entry-type identifies what kind of atomic unit this entry is
	"entry-type": #EntryType @go(EntryType)
}

// EntryType enumerates the atomic units within Gemara artifacts that can participate in mappings
#EntryType: "Guideline" | "Statement" | "Control" | "AssessmentRequirement" | "Capability" | "Threat" | "Risk" | "Vector" @go(-)
