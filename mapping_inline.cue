// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")

package gemara

// MappingReference represents a reference to an external document with full metadata.
// Fields constrained by #URL and #Digest pin their Go projection to string. A named
// Go type here buys no validation — Go has no constructor to enforce the pattern —
// and would cost every consumer a conversion at every use site. The constraint is
// the value of the definition; the Go name is not.

// URL is a URI with a scheme. Any scheme is accepted (e.g. https, file, oci, s3,
// arn), so content hosted outside http(s) can be addressed.
#URL: =~"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]+$" @go(-)

// Digest is a cryptographic hash of a full octet stream; format: algorithm:encoded
// (e.g. sha256:<64 lowercase hex>). sha256 is the MUST-support floor for every
// conforming verifier; sha512 and blake3 are MAY. Registered algorithms are
// additionally checked for that algorithm's encoding and length; any other algorithm
// is checked for grammar only, per the OCI rule that unrecognized algorithms
// complying with the grammar pass. That is how the algorithm set extends without
// this field becoming a closed enum.
//
// The hash covers the full octet stream as delivered — never a canonical
// re-serialization, and never a sub-resource selected by coordinate or entry-id,
// which are reader hints rather than inputs to the hash. A verifier reports exactly
// one of three outcomes per citation: verified, integrity-failure, or unverifiable
// (retrieval failed, or the algorithm is unregistered or unimplemented). None of
// them may be reported as valid. Absence of digest means no integrity claim was
// made, not that the content is known unchanged.
//
// It is one digest rather than a set of them. The algorithm travels in the value, so
// adopting a new one needs no schema change, and the MUST-support floor means the
// digest a citation carries is one every conforming verifier can check — which is
// all that ISO 19011's requirement that audit evidence be verifiable asks for. A set
// would buy only simultaneous publication for verifiers of differing capability,
// which the floor already removes, at the cost of a rule this schema cannot enforce:
// a set holding one matching and one mismatching digest has two defensible readings,
// and CUE cannot hash. Handing a subject or materials list to an in-toto verifier
// would need that type's DigestSet; the place for it then is a content descriptor
// primitive, not a wider #Digest, whose projections every consumer reads.
#Digest: (=~"^[a-z0-9]+(?:[+._-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$" &
	(=~"^(?:sha256:[a-f0-9]{64}|sha512:[a-f0-9]{128}|blake3:[a-f0-9]{64})$" |
	!~"^(?:sha256|sha512|blake3):")) @go(-)

#MappingReference: {
	// id identifies this mapping reference within the artifact and, when url
	// is absent, the referenced artifact's metadata.id.
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
	url?: =~"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]+$"
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

// EvidenceMapping identifies the source from which evidence was collected.
// reference-id names the MappingReference; coordinate or entry-id gives
// specificity within it; digest pins the observed content at collection time.
#EvidenceMapping: {
	// reference-id ties this evidence to a mapping-reference in the artifact's metadata
	"reference-id": string @go(ReferenceId)

	// coordinate is the precise location within the stream identified by reference-id
	// (e.g. an API path, file path, or JSON path expression). May be combined with
	// entry-id to identify a sub-location within that entry's output. It is a reader
	// hint, not an address a verifier resolves, and not an input to digest.
	coordinate?: string

	// entry-id identifies a specific entry within a referenced Gemara artifact. May be
	// combined with coordinate. Like coordinate, it is a reader hint rather than an
	// input to digest: the hash covers the whole stream, not the entry selected from it.
	"entry-id"?: string @go(EntryId)

	// download-url is the address this content can be retrieved from, where that is
	// not the url of the mapping-reference named by reference-id.
	DU="download-url"?: #URL @go(DownloadUrl,type=string)

	// digest is a cryptographic hash of the full octet stream this citation refers to,
	// as delivered. See #Digest for what a verifier must do with it, and for why
	// absence is not assurance.
	digest?: #Digest @go(Digest,type=string)

	// size is the length, in bytes, of that octet stream.
	size?: int & >=0

	// media-type is the IANA media type of that content, e.g. application/json,
	// text/yaml.
	"media-type"?: =~"^[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*/[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*$" @go(MediaType)

	// remarks is prose regarding this evidence reference
	remarks?: string

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	// A download-url says a verifier can fetch this content, so it comes with the
	// digest of what was fetched: whoever retrieved the bytes could hash them, and an
	// address with no digest says where the artifact is without saying what it was.
	// A digest without a download-url stays legal, because reference-id resolves to a
	// mapping-reference that carries the address.
	//
	// Stated on digest, the field it requires, rather than on download-url, the one it
	// would forbid: an error() on download-url puts bottom in that field's type, and
	// the Go projection degrades the field to `any` for every consumer.

	if DU != _|_ {
		digest: #Digest
	}
}
