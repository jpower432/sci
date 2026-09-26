// SPDX-License-Identifier: Apache-2.0

// Schema lifecycle: experimental | stable | deprecated
@gemara(status="stable")
package gemara

import "list"

@go(gemara)

// Catalog describes a set of topically-associated entries
#Catalog: {
	// title describes the purpose of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// groups contains a list of groups that can be referenced by entries in this catalog
	groups?: [#Group, ...#Group]

	// extends references catalogs that this catalog builds upon
	extends?: [...#ArtifactMapping] @go(Extends)

	imports?: [#MultiEntryMapping, ...#MultiEntryMapping]

	// ---- Validation --------------------------------------------------------
	// Comments in validation sections stay detached (blank line after), so they
	// are never published as a field's API description.

	if groups != _|_ {
		_uniqueGroupIds: {for i, g in groups {(g.id): i}}
	}

	// A catalog that extends or imports declares the references it names, and each
	// reference-id names one of them: the same rule every log applies to its own
	// pointers, which a catalog had only half of — it required the declarations to
	// exist without checking that anything matched.

	_catalogRefIds: list.Concat([[metadata.id], [for r in metadata."mapping-references" {r.id}]])

	if extends != _|_ {
		metadata: "mapping-references": [#MappingReference, ...#MappingReference]

		for i, e in extends {
			_refValidation: "extends-\(i)": _catalogRefIds & list.Contains(e."reference-id")
		}
	}

	if imports != _|_ {
		metadata: "mapping-references": [#MappingReference, ...#MappingReference]

		for i, m in imports {
			_refValidation: "imports-\(i)": _catalogRefIds & list.Contains(m."reference-id")
		}
	}
}

// Log describes a set of recorded entries from a measurement activity
#Log: {
	// metadata provides detailed data about this log
	metadata: #Metadata @go(Metadata)

	// target identifies the resource being evaluated
	target: #Resource @go(Target)

	// ---- Validation --------------------------------------------------------

	// _refIds is every id a reference-id in this log may name: its declared
	// mapping-references, plus its own metadata.id for a self-reference.
	// Computed once here so each log's validation reads it rather than
	// rebuilding it. It lives on #Log, not #Metadata: adding it to #Metadata
	// changes the generated Go type of every artifact that constrains
	// metadata.mapping-references.

	if metadata."mapping-references" != _|_ {
		_refIds: list.Concat([[metadata.id], [for r in metadata."mapping-references" {r.id}]])
	}
	if metadata."mapping-references" == _|_ {_refIds: [metadata.id]}
}

// Lifecycle represents the lifecycle state of a guideline, control, or assessment requirement
#Lifecycle: *"Active" | "Draft" | "Deprecated" | "Retired" @go(-)

// ConfidenceLevel indicates the evaluator's confidence level in an assessment result.
#ConfidenceLevel: "Undetermined" | "Low" | "Medium" | "High" @go(-)
