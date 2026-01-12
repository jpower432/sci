package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func main() {
	lexiconPath := flag.String("lexicon", "lexicon.yaml", "path to lexicon YAML file")
	outputPath := flag.String("output", "", "path to output markdown file (required)")
	version := flag.String("version", "dev", "version string for the lexicon")
	flag.Parse()

	if *outputPath == "" {
		fmt.Fprintf(os.Stderr, "Error: output path is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if err := compileLexicon(*lexiconPath, *outputPath, *version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Lexicon compiled to %s\n", *outputPath)
}

func compileLexicon(lexiconPath, outputPath, version string) error {
	data, err := os.ReadFile(lexiconPath)
	if err != nil {
		return fmt.Errorf("reading lxn file: %w", err)
	}

	var lxn Lexicon
	if err := yaml.Unmarshal(data, &lxn); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	if version != "" {
		lxn.Version = version
	}

	var buf bytes.Buffer
	if err := lxn.ToMarkdownPage(&buf); err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}
