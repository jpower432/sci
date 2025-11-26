package common

// ============================================================================
// Mapping Types - MappingReference, MappingEntry, Mapping
// ============================================================================

// MappingReference represents a reference to an external document.
#MappingReference: {
	id:           string
	title:        string
	version:      string
	description?: string
	url?:         =~"^(https?|file)://[^\\s]+$"
}

// Mapping represents a mapping to an external reference with one or more entries.
#Mapping: {
	// ReferenceId should reference the corresponding MappingReference id from metadata
	"reference-id": string @go(ReferenceId)
	entries: [...#MappingEntry] @go(Entries)
	remarks?: string
}

// SimpleMapping represents a simple mapping with exactly one entry and optional reference-id
#SimpleMapping: {
	// ReferenceId should reference the corresponding MappingReference id from metadata
	"reference-id"?: string @go(ReferenceId)
	"entry-id":      string @go(EntryId)
	remarks?:        string
}

// MappingEntry represents a single entry within a mapping
#MappingEntry: {
	"reference-id": string @go(ReferenceId)
	strength:       int & >=1 & <=10
	remarks?:       string
}
