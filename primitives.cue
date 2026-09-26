// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")

package gemara

// MappingReference represents a reference to an external document with full metadata.
#MappingReference: {
	// id identifies this mapping reference within the artifact; every reference-id
	// in the artifact MUST equal one of these ids or the artifact's own metadata.id.
	id: string

	// title describes the purpose of this mapping reference at a glance
	title: string

	// version is the version identifier of the artifact being mapped to
	version: string

	// description is prose regarding the artifact's purpose or content
	description?: string

	// url is the path where the artifact may be retrieved; preferably responds with Gemara-compatible YAML/JSON.
	// Any URI scheme is accepted (e.g. https, file, oci, s3, arn) so evidence can be
	// addressed wherever it actually lives.
	url?: #URL
}

// ArtifactMapping represents a mapping to an external artifact or artifact entry
#ArtifactMapping: {
	// reference-id identifies an element from a MappingReference in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// remarks is prose regarding the mapped artifact or the mapping relationship
	remarks?: string
}

// MultiEntryMapping represents a mapping to an external reference with one or more entries.
#MultiEntryMapping: {
	// top-level reference to the MappingReference entry
	#ArtifactMapping

	// entries is a list of mapping entries
	entries: [#ArtifactMapping, ...#ArtifactMapping] @go(Entries)
}

// EntryMapping represents how a specific entry maps to a MappingReference.
#EntryMapping: {
	// reference-id is the id for a MappingReference entry in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// entry-id is the identifier being mapped to in the referenced artifact
	"entry-id": string @go(EntryId)

	// remarks is prose describing the mapping relationship
	remarks?: string
}

// URL validates an absolute URI with any scheme (e.g. https, file, oci, s3).
#URL: =~"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]+$" @go(-)

// Digest is a cryptographic hash of a full octet stream; format: algorithm:encoded
// (e.g. sha256:<64 lowercase hex>). sha256 is the MUST-support floor for every
// conforming verifier; sha512 and blake3 are MAY. Registered algorithms are
// additionally checked for that algorithm's encoding and length; any other
// algorithm is checked for grammar only, per the OCI rule that unrecognized
// algorithms complying with the grammar pass.
//
// The hash covers the full octet stream retrieved from download-url exactly
// as delivered. A verifier reports exactly one of three outcomes per citation
// — verified, integrity-failure, or unverifiable (no digest, no download-url, retrieval
// failed, or the algorithm is unregistered/unimplemented).
// Absence of digest means no integrity claim was made, not that the content is
// known unchanged.
//
// It is one digest rather than a set of them. The algorithm travels in the value,
// so adopting a new one needs no schema change, and the MUST-support floor means
// the digest a citation carries is one every conforming verifier can check —
// which is all that ISO 19011's requirement that audit evidence be verifiable
// asks for. A set would buy only simultaneous publication for verifiers of
// differing capability, which the floor already removes, at the cost of a rule
// this schema cannot enforce: a set holding one matching and one mismatching
// digest has two defensible readings, and CUE cannot hash. Handing a subject or
// materials list to an in-toto verifier would need that type's DigestSet; the
// place for it then is a content descriptor primitive, not a wider #Digest,
// whose Go and OpenAPI projections every consumer reads.
#Digest: (=~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$" &
	(=~"^(?:sha256:[a-f0-9]{64}|sha512:[a-f0-9]{128}|blake3:[a-f0-9]{64})$" |
	!~"^(?:sha256|sha512|blake3):")) @go(-)
