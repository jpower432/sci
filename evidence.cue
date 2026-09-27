// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

@go(gemara)

// Evidence records what was cited to support an opinion for a specific activity:
// raw data for the evaluation layer, evaluation and enforcement artifacts for the audit layer.
// At least one of payload or source MUST be present; an entry with neither is semantically incomplete.
// payload carries decoded evidence inline and is not digested; integrity for
// referenced evidence is expressed by source.digest over source.download-url,
// via #ContentDescriptor.
#Evidence: {
	// id uniquely identifies this evidence
	id: string

	// type categorizes the kind of evidence
	type: #EvidenceType

	// collected-at is the timestamp when the evidence was gathered
	"collected-at": #Datetime @go(CollectedAt)

	// originator is the party that produced the evidence content.
	originator?: #Actor

	// collector is the party that gathered this evidence into the log, recorded
	// only when it differs from the log's metadata.author. Third-party evidence
	// has a real distinction to draw — a scanner produced it, an auditor pulled it
	// in — but where the log's author gathered it, saying so again would be a
	// second copy of the same fact.
	collector?: #Actor

	// payload is the raw evidence data collected inline
	payload?: _ @go(Payload,type=any)

	// source identifies the artifact or system from which this evidence was collected and
	// retrieval information.
	source?: #EvidenceMapping @go(Source)

	// description explains what this evidence represents
	description?: string
}

// EvidenceType categorizes the kind of evidence. It remains an open enum:
// recommended values include artifact types already known to Gemara (e.g.
// EvaluationLog, EnforcementLog) plus categories for common evidence forms.
#EvidenceType: #ArtifactType | #URL @go(-)

// EvidenceMapping identifies the source from which evidence was collected.
// reference-id names the MappingReference; coordinate and entry-id are reader
// hints for locating content within it. Content addressing and integrity
// (download-url, digest, size, media-type) come from #ContentDescriptor.
#EvidenceMapping: {
	// reference-id defines the evidence artifact source's identifying information.
	"reference-id": string @go(ReferenceId)

	// coordinate is the precise location within the stream identified by reference-id
	// (e.g. an API path, file path, or JSON path expression). May be combined with
	// entry-id. It is a reader hint, not an address a verifier resolves, and not an
	// input to digest.
	coordinate?: string

	// entry-id identifies a specific entry within a referenced Gemara artifact.
	// May be combined with coordinate. It is a reader hint, not an address a
	// verifier resolves, and not an input to digest.
	"entry-id"?: string @go(EntryId)

	// Embedded information about how evidence can be retrieved.
	// download-url is the address this content can be retrieved from is it not
	// available from the MappingReference.url.
	DU="download-url"?: #URL @go(DownloadUrl,type=string)

	// digest is a cryptographic hash of the full octet stream retrieved from
	// download-url. See #Digest for what a verifier must do with it.
	digest?: #Digest @go(Digest,type=string)

	// size is the length, in bytes, of the octet stream at download-url.
	size?: int & >=0

	// media-type is the IANA media type of the content at download-url,
	// e.g. application/json, text/yaml.
	"media-type"?: =~"^[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*/[a-zA-Z0-9][a-zA-Z0-9!#$&.+^_-]*$" @go(MediaType)

	// remarks is prose regarding this evidence reference
	remarks?: string

	// ---- Validation --------------------------------------------------------

	// A download-url says a verifier can fetch this content, so it comes with the
	// digest of what was fetched: whoever retrieved the bytes could hash them, and
	// an address without a digest says where the artifact is without saying what it
	// was. reference-id and coordinate already cover "here is where to look"
	// without claiming retrievability.
	//
	// Stated on digest, the field it requires, rather than on download-url, the one
	// it would forbid: an error() on download-url puts bottom in that field's type,
	// and the Go projection degrades the field to `any` for every consumer.

	if DU != _|_ {
		digest: #Digest
	}
}

// ---- Validation ------------------------------------------------------------
// #_EvidenceStrict carries every rule on #Evidence. It is hidden (@go(-), and
// never projected), so the structures above stay shape and documentation only.

// _EvidenceStrict layers the "at least one of payload or source" rule on top of #Evidence
#_EvidenceStrict: {
	@go(-)
} & #Evidence & {
	// payload is the raw evidence data collected inline
	payload?: _

	if payload == _|_ {
		source: #EvidenceMapping
	}

	// An inline payload and a source download-url are mutually exclusive:
	// inline content has no address to retrieve it from, and content with an
	// address is referenced rather than carried. source may still name where an
	// inline payload came from, by reference-id, without a retrievable address.
	if payload != _|_ {
		source?: "download-url"?: error("an inline payload cannot also have a source download-url: inline content is carried, referenced content is addressed")
	}
}
