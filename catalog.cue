// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// Catalog describes a set of topically-associated entries
#Catalog: {
	// title describes the purpose of this catalog at a glance
	title: string

	// metadata provides detailed data about this catalog
	metadata: #Metadata @go(Metadata)

	// extends references catalogs that this catalog builds upon
	extends?: [...#ArtifactMapping] @go(Extends)

	imports?: [#MultiEntryMapping, ...#MultiEntryMapping]
}
