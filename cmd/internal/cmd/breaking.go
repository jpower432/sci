// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// Change is an ERR-level backward-incompatible change reported by oasdiff.
type Change struct {
	ID   string
	Text string
}

// loadWrapped loads an OpenAPI document and synthesizes bidirectional paths for
// every component schema so oasdiff (which diffs the API surface, not bare
// component schemas) can classify schema-level changes. Each schema is exposed
// as a GET response body (consumer direction: removed required property =
// breaking) and a POST request body (producer direction: new required property =
// breaking).
func loadWrapped(path string) (*openapi3.T, error) {
	doc, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	if doc.Paths == nil {
		doc.Paths = openapi3.NewPaths()
	}
	for name := range doc.Components.Schemas {
		ref := doc.Components.Schemas[name] // already-resolved SchemaRef
		mt := openapi3.NewContentWithJSONSchemaRef(ref)
		emptyOK := openapi3.NewResponses(openapi3.WithStatus(200,
			&openapi3.ResponseRef{Value: openapi3.NewResponse().WithDescription("ok")}))
		doc.Paths.Set("/_schema/"+name, &openapi3.PathItem{
			Get: &openapi3.Operation{
				OperationID: "get_" + name,
				Responses: openapi3.NewResponses(openapi3.WithStatus(200,
					&openapi3.ResponseRef{Value: openapi3.NewResponse().WithDescription("ok").WithContent(mt)})),
			},
			Post: &openapi3.Operation{
				OperationID: "post_" + name,
				RequestBody: &openapi3.RequestBodyRef{
					Value: openapi3.NewRequestBody().WithRequired(true).WithContent(mt),
				},
				Responses: emptyOK,
			},
		})
	}
	return doc, nil
}

// breakingChanges returns the ERR-level backward-incompatible changes from base
// to rev. Both documents must already be wrapped via loadWrapped.
func breakingChanges(base, rev *openapi3.T) ([]Change, error) {
	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(),
		&load.SpecInfo{Url: "base", Spec: base},
		&load.SpecInfo{Url: "rev", Spec: rev})
	if err != nil {
		return nil, err
	}
	loc := checker.NewDefaultLocalizer()
	var out []Change
	for _, c := range checker.CheckBackwardCompatibility(checker.NewConfig(checker.GetAllChecks()), d, sources) {
		if c.GetLevel() == checker.ERR {
			out = append(out, Change{ID: c.GetId(), Text: c.GetUncolorizedText(loc)})
		}
	}
	return out, nil
}
