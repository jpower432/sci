package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"reflect"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run utils/fix_generated_types.go <filename>")
		return
	}

	filename := os.Args[1]
	src, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Fix import paths
	// Replace "github.com/ossf/gemara/schemas/common" with "github.com/ossf/gemara/common"
	// Replace "github.com/ossf/gemara/schemas/layer1" with "github.com/ossf/gemara/layer1"
	// Replace "github.com/ossf/gemara/schemas/layer2" with "github.com/ossf/gemara/layer2"
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			if x.Path != nil {
				importPath := strings.Trim(x.Path.Value, `"`)
				switch importPath {
				case "github.com/ossf/gemara/schemas/common":
					x.Path.Value = `"github.com/ossf/gemara/common"`
				case "github.com/ossf/gemara/schemas/layer1":
					x.Path.Value = `"github.com/ossf/gemara/layer1"`
				case "github.com/ossf/gemara/schemas/layer2":
					x.Path.Value = `"github.com/ossf/gemara/layer2"`
				}
			}
		}
		return true
	})

	// Add YAML tags to struct fields that have JSON tags but no YAML tags
	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range st.Fields.List {
			if field.Tag == nil {
				continue
			}
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			jsonTag := tag.Get("json")
			if jsonTag == "" {
				continue
			}
			if tag.Get("yaml") != "" {
				continue // already has yaml tag
			}
			// Compose new tag string
			var tagParts []string
			for _, part := range strings.Split(string(tag), " ") {
				if strings.HasPrefix(part, "yaml:") {
					continue
				}
				tagParts = append(tagParts, part)
			}
			tagParts = append(tagParts, fmt.Sprintf(`yaml:"%s"`, jsonTag))
			newTag := "`" + strings.Join(tagParts, " ") + "`"
			field.Tag.Value = newTag
		}
		return true
	})

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		panic(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filename, formatted, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Fixed imports and added YAML tags to", filename)
}
