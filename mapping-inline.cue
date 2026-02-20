// Schema lifecycle: experimental | stable | deprecated
@status("stable")

package gemara

// ArtifactMapping represents a reference to an external artifact with remarks
#ArtifactMapping: {
	// reference-id should reference the corresponding MappingReference id from metadata
	"reference-id": string @go(ReferenceId)

	// remarks is prose regarding the mapped artifact or the mapping relationship
	"remarks": string
}

// MultiEntryMapping represents a mapping to an external reference with one or more entries.
#MultiEntryMapping: {
	// reference-id should reference the corresponding MappingReference id from metadata
	"reference-id": string @go(ReferenceId)

	// entries is a list of mapping entries
	entries: [#MappingEntry, ...#MappingEntry] @go(Entries)

	// remarks is prose regarding the mapped artifact or the mapping relationship
	remarks?: string
}

// EntryMapping represents how a specific entry (control/requirement/procedure) maps to a MappingReference.
#EntryMapping: {
	// reference-id is the id for a MappingReference entry in the artifact's metadata
	"reference-id"?: string @go(ReferenceId)

	// entry-id is the identifier being mapped to in the referenced artifact
	"entry-id": string @go(EntryId)

	// strength is the author's estimate of how completely the current/source material satisfies the target/reference material;
	// Range: 1-10. Zero value means not yet quantified.
	strength?: int & >=1 & <=10

	// remarks is prose describing the mapping relationship
	remarks?: string
}

// MappingEntry represents a single entry within a mapping
#MappingEntry: {
	// reference-id is the id for a MappingReference entry in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// strength is the author's estimate of how completely the current/source material satisfies the target/reference material;
	// Range: 1-10. Zero value means not yet quantified.
	strength?: int & >=1 & <=10

	// remarks is prose describing the mapping relationship
	remarks?: string
}
