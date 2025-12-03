package layer1

import (
	"fmt"
	"path"

	"github.com/ossf/gemara/internal/loaders"
)

// LoadFiles loads data from any number of YAML or JSON files at the provided paths.
// sourcePath are expected to be file or https URIs in the form file:///path/to/file.yaml or https://example.com/file.yaml.
// If run multiple times, this method will append new data to previous data.
func (g *Guidance) LoadFiles(sourcePaths []string) error {
	for _, sourcePath := range sourcePaths {
		doc := &Guidance{}
		if err := doc.LoadFile(sourcePath); err != nil {
			return err
		}
		if g.Title == "" {
			g.Title = doc.Title
			g.DocumentType = doc.DocumentType
			g.Exemptions = doc.Exemptions
			g.FrontMatter = doc.FrontMatter
			g.Metadata = doc.Metadata
		}
		g.Families = append(g.Families, doc.Families...)
		g.Guidelines = append(g.Guidelines, doc.Guidelines...)
	}
	return nil
}

// LoadFile loads data from a YAML or JSON file at the provided path into the Guidance.
// sourcePath is expected to be a file or https URI in the form file:///path/to/file.yaml or https://example.com/file.yaml.
// If run multiple times for the same data type, this method will override previous data.
func (g *Guidance) LoadFile(sourcePath string) error {
	ext := path.Ext(sourcePath)
	switch ext {
	case ".yaml", ".yml":
		err := loaders.LoadYAML(sourcePath, g)
		if err != nil {
			return err
		}
	case ".json":
		err := loaders.LoadJSON(sourcePath, g)
		if err != nil {
			return fmt.Errorf("error loading json: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
	return nil
}
