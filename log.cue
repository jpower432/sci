// Schema lifecycle: experimental | stable | deprecated
@status("experimental")
package gemara

@go(gemara)

// Log describes a set of recorded entries from a measurement activity
#Log: {
	// metadata provides detailed data about this log
	metadata: #Metadata @go(Metadata)

	// target identifies the resource being evaluated
	target: #Resource @go(Target)
}
