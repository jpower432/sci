package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

type Term struct {
	Term       string   `yaml:"term"`
	Definition string   `yaml:"definition"`
	References []string `yaml:"references"`
}

type TemplateData struct {
	Table string
}

var lexicon2MDCmd = &cobra.Command{
	Use:   "lexicon2md",
	Short: "Generate definitions table from lexicon YAML",
	RunE:  runLexicon2MD,
}

var lexicon2MDFlags struct {
	lexiconFile string
	outputFile  string
}

func newLexicon2MDCmd() *cobra.Command {
	lexicon2MDCmd.Flags().StringVarP(&lexicon2MDFlags.lexiconFile, "lexicon", "l", "docs/lexicon.yaml", "Input lexicon YAML file")
	lexicon2MDCmd.Flags().StringVarP(&lexicon2MDFlags.outputFile, "output", "o", "docs/model/02-definitions.md", "Output markdown file")
	return lexicon2MDCmd
}

func runLexicon2MD(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(lexicon2MDFlags.lexiconFile)
	if err != nil {
		return fmt.Errorf("Error reading lexicon file: %v", err)
	}

	var terms []Term
	if err := yaml.Unmarshal(data, &terms); err != nil {
		return fmt.Errorf("Error parsing lexicon YAML: %v", err)
	}

	var tableRows strings.Builder
	for _, term := range terms {
		slug := termToSlug(term.Term)

		termName := fmt.Sprintf("<a id=\"%s\"></a>**%s**", slug, term.Term)

		layer := ""
		if len(term.References) > 0 {
			layer = term.References[0]
		}

		definition := strings.ReplaceAll(term.Definition, "|", "\\|")

		tableRows.WriteString(fmt.Sprintf("| %s | %s | %s |\n", termName, definition, layer))
	}

	templateContent, err := os.ReadFile(lexicon2MDFlags.outputFile)
	if err != nil {
		return fmt.Errorf("Error reading output file: %v", err)
	}

	tmpl, err := template.New("definitions").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("Error parsing template: %v", err)
	}

	var output bytes.Buffer
	templateData := TemplateData{
		Table: strings.TrimSpace(tableRows.String()),
	}
	if err := tmpl.Execute(&output, templateData); err != nil {
		return fmt.Errorf("Error executing template: %v", err)
	}

	if err := os.WriteFile(lexicon2MDFlags.outputFile, output.Bytes(), 0644); err != nil {
		return fmt.Errorf("Error writing output file: %v", err)
	}

	fmt.Printf("Successfully generated definitions table in %s\n", lexicon2MDFlags.outputFile)
	return nil
}

func termToSlug(term string) string {
	slug := strings.ToLower(term)
	slug = strings.ReplaceAll(slug, " ", "-")
	var result strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
