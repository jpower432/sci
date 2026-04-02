// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
	"golang.org/x/mod/semver"
)

const modulePath = "github.com/gemaraproj/gemara"

// stableArtifactDefs lists the CUE definitions that have reached stable status
// via @status("stable") in their respective .cue files.
// Update this list when promoting a schema to stable.
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

// TestNoBreakingChanges pulls the latest release from the CUE registry and
// verifies that each stable schema definition remains backward compatible.
//
// By default only stable releases are considered. Set GEMARA_COMPAT_PRERELEASE=true
// to also include pre-release versions, which is useful before v1.0.0 is published.
func TestNoBreakingChanges(t *testing.T) {
	ctx := context.Background()

	resolver, err := modconfig.NewResolver(&modconfig.Config{
		CUERegistry: modconfig.DefaultRegistry,
	})
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}
	regClient := modregistry.NewClientWithResolver(resolver)

	includePrerelease := os.Getenv("GEMARA_COMPAT_PRERELEASE") == "true"

	latestVer, err := latestVersion(ctx, regClient, modulePath, includePrerelease)
	if err != nil {
		t.Logf("no suitable release found | skipping compatibility check; %v", err)
		t.Skip()
	}
	t.Logf("comparing against released version: %s", latestVer)

	reg, err := modconfig.NewRegistry(&modconfig.Config{
		CUERegistry: modconfig.DefaultRegistry,
	})
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	oldSchema, err := loadModuleFromRegistry(reg, latestVer)
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

// latestVersion returns the most recent version of the module from the registry.
// If includePrerelease is false, only stable releases are considered.
func latestVersion(ctx context.Context, client *modregistry.Client, modPath string, includePrerelease bool) (module.Version, error) {
	versions, err := client.ModuleVersions(ctx, modPath+"@v1")
	if err != nil {
		return module.Version{}, fmt.Errorf("listing versions for %s: %w", modPath, err)
	}
	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i]
		if includePrerelease || semver.Prerelease(v) == "" {
			return module.NewVersion(modPath, v)
		}
	}
	if includePrerelease {
		return module.Version{}, fmt.Errorf("no versions found for %s", modPath)
	}
	return module.Version{}, fmt.Errorf("no stable release found for %s (set GEMARA_COMPAT_PRERELEASE=true to include pre-releases)", modPath)
}

// loadModuleFromRegistry fetches the given module version from the registry
// and builds a CUE value using load.Instances, which correctly handles
// cross-file references within the module.
func loadModuleFromRegistry(reg modconfig.Registry, ver module.Version) (cue.Value, error) {
	instances := load.Instances([]string{ver.String()}, &load.Config{
		Registry: reg,
	})
	if len(instances) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE instances returned for %v", ver)
	}
	if err := instances[0].Err; err != nil {
		return cue.Value{}, fmt.Errorf("loading module %v: %w", ver, err)
	}
	val := schemaCtx.BuildInstance(instances[0])
	if err := val.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("building schema for %v: %w", ver, err)
	}
	return val, nil
}
