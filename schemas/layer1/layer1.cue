package layer1

import "github.com/ossf/gemara/schemas/common"

@go(layer1)

// Guidance represents a Layer 1 guidance document
#Guidance: {
	metadata?:       common.#Metadata @go(Metadata)
	title:           string
	"document-type": #DocumentType @go(DocumentType) @yaml("document-type")
	"front-matter"?: string        @go(FrontMatter) @yaml("front-matter,omitempty")
	exemptions?: [...#Exemption] @go(Exemptions)
	families: [...common.#Family] @go(Families)
	guidelines: [...#Guideline] @go(Guidelines)
}

// Exemption represents an exemption with a reason and optional redirect
#Exemption: {
	reason:    string
	redirect?: common.#Mapping @go(Redirect)
}

#DocumentType: "Standard" | "Regulation" | "Best Practice" | "Framework"

// Guideline represents a single guideline within a guidance document
#Guideline: {
	id:           string
	title:        string
	objective?:   string
	"family-id"?: string @go(FamilyId) @yaml("family-id,omitempty")
	recommendations?: [...string]
	// Extends allows you to add supplemental guidance within a local guidance document
	// like a control enhancement or from an imported guidance document.
	extends?:   common.#SimpleMapping @go(Extends)
	rationale?: #Rationale            @go(Rationale,optional=nillable)
	statements?: [...#Statement] @go(Statements)
	"guideline-mappings"?: [...common.#Mapping] @go(GuidelineMappings) @yaml("guideline-mappings,omitempty")
	"principle-mappings"?: [...common.#Mapping] @go(PrincipleMappings) @yaml("principle-mappings,omitempty")
	"see-also"?: [...common.#SimpleMapping] @go(SeeAlso) @yaml("see-also,omitempty")
}

// Statement represents a sub-statement within a guideline
#Statement: {
	id:     string
	title?: string
	text:   string
	recommendations?: [...string]
}

// Rationale provides contextual information to help with development and understanding of
// guideline intent.
#Rationale: {
	importance: string
	goals: [...string]
}
