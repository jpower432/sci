package main

import (
	"fmt"
	"os"

	"github.com/ossf/gemara/layer3"
)

func main() {
	policyFile := "osps-example-policy.yaml"
	if len(os.Args) >= 2 {
		policyFile = os.Args[1]
	}

	fmt.Fprintf(os.Stderr, "Loading policy from: %s\n", policyFile)

	// Load the policy
	policy := &layer3.Policy{}
	// Add file:// prefix if not present
	loadPath := policyFile
	if policyFile[:7] != "file://" {
		loadPath = "file://" + policyFile
	}
	if err := policy.LoadFile(loadPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading policy: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Policy loaded: %s (v%s)\n", policy.Metadata.Id, policy.Metadata.Version)
	fmt.Fprintf(os.Stderr, "Generating checklist...\n\n")

	// Generate markdown checklist
	markdown, err := policy.ToMarkdownChecklist()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating checklist: %v\n", err)
		os.Exit(1)
	}

	// Print the checklist
	fmt.Println(markdown)
}
