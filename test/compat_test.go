// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
	"golang.org/x/mod/semver"
)

const modulePath = "github.com/gemaraproj/gemara"

var stableArtifactDefs = []string{
	"#ControlCatalog",
	"#CapabilityCatalog",
	"#EvaluationLog",
	"#ThreatCatalog",
	"#Metadata",
	"#ArtifactType",
	"#Group",
	"#Datetime",
}

func TestNoBreakingChanges(t *testing.T) {
	ctx := context.Background()

	resolver, err := modconfig.NewResolver(&modconfig.Config{
		CUERegistry: modconfig.DefaultRegistry,
	})
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}
	regClient := modregistry.NewClientWithResolver(resolver)

	latestVer, err := latestReleasedVersion(ctx, regClient, modulePath)
	if err != nil {
		t.Logf("no stable release found — latest published version is a pre-release, skipping compatibility check")
		t.Skip()
	}
	t.Logf("comparing against released version: %s", latestVer)

	oldSchema, err := loadModuleFromRegistry(ctx, regClient, latestVer)
	if err != nil {
		t.Fatalf("failed to load released module: %v", err)
	}

	for _, defPath := range stableArtifactDefs {
		defPath := defPath
		t.Run(defPath, func(t *testing.T) {
			newDef := schemaValue.LookupPath(cue.ParsePath(defPath))
			if newDef.Err() != nil {
				t.Fatalf("new schema: lookup %s: %v", defPath, newDef.Err())
			}

			oldDef := oldSchema.LookupPath(cue.ParsePath(defPath))
			if oldDef.Err() != nil {
				t.Logf("definition %s not found in released version (new addition)", defPath)
				return
			}

			t.Logf("validating %s: checking new definition subsumes old definition", defPath)
			if err := newDef.Subsume(oldDef, cue.Raw(), cue.Schema()); err != nil {
				t.Errorf("breaking change detected in %s:\n%v", defPath, err)
			} else {
				t.Logf("%s: OK — no breaking changes", defPath)
			}
		})
	}
}

func latestReleasedVersion(ctx context.Context, client *modregistry.Client, modPath string) (module.Version, error) {
	versions, err := client.ModuleVersions(ctx, modPath+"@v1")
	if err != nil {
		return module.Version{}, fmt.Errorf("listing versions for %s: %w", modPath, err)
	}
	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i]
		if semver.Prerelease(v) == "" {
			return module.NewVersion(modPath, v)
		}
	}
	return module.Version{}, fmt.Errorf("no stable release found for %s", modPath)
}

func loadModuleFromRegistry(ctx context.Context, client *modregistry.Client, ver module.Version) (cue.Value, error) {
	mod, err := client.GetModule(ctx, ver)
	if err != nil {
		return cue.Value{}, fmt.Errorf("get module %v: %w", ver, err)
	}

	zipRC, err := mod.GetZip(ctx)
	if err != nil {
		return cue.Value{}, fmt.Errorf("get zip for %v: %w", ver, err)
	}
	defer zipRC.Close()

	zipBytes, err := io.ReadAll(zipRC)
	if err != nil {
		return cue.Value{}, fmt.Errorf("read zip for %v: %w", ver, err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return cue.Value{}, fmt.Errorf("open zip for %v: %w", ver, err)
	}

	cueCtx := cuecontext.New()
	val := cueCtx.CompileBytes([]byte(""))

	for _, f := range zipReader.File {
		if !strings.HasSuffix(f.Name, ".cue") || strings.Contains(f.Name, "cue.mod/") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return cue.Value{}, fmt.Errorf("open %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return cue.Value{}, fmt.Errorf("read %s: %w", f.Name, err)
		}
		compiled := cueCtx.CompileBytes(data, cue.Filename(f.Name))
		if compiled.Err() != nil {
			return cue.Value{}, fmt.Errorf("compile %s: %w", f.Name, compiled.Err())
		}
		val = val.Unify(compiled)
	}
	return val, nil
}
