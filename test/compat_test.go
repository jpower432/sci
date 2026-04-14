// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/mod/modconfig"
	"cuelang.org/go/mod/modregistry"
	"cuelang.org/go/mod/module"
	"golang.org/x/mod/semver"
)

const modulePath = "github.com/gemaraproj/gemara"

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
		t.Logf("no suitable release found | skipping compatibility check: %v", err)
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

	schemaDir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to resolve schema directory: %v", err)
	}

	stableDefs, err := collectStableDefs(schemaDir)
	if err != nil {
		t.Fatalf("failed to collect stable definitions: %v", err)
	}
	t.Logf("found %d stable definitions to check", len(stableDefs))

	for _, defPath := range stableDefs {
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

			assertFieldsCompatible(t, defPath, newDef, oldDef)
		})
	}
}

// assertFieldsCompatible checks that the new definition hasn't removed any
// fields present in the old definition and hasn't introduced new required
// fields. CUE's Subsume API produces false positives when comparing values
// built from different load contexts (local filesystem vs OCI registry), so
// we compare field structure directly using name-based maps.
func assertFieldsCompatible(t *testing.T, path string, newDef, oldDef cue.Value) {
	t.Helper()

	oldFields := fieldNames(oldDef, true)
	newAllFields := fieldNames(newDef, true)
	newRequiredFields := fieldNames(newDef, false)

	for _, name := range oldFields {
		if !contains(newAllFields, name) {
			t.Errorf("field %s removed from %s", name, path)
		}
	}

	oldSet := make(map[string]bool, len(oldFields))
	for _, name := range oldFields {
		oldSet[name] = true
	}
	for _, name := range newRequiredFields {
		if !oldSet[name] {
			t.Errorf("new required field %s added to %s", name, path)
		}
	}
}

func fieldNames(v cue.Value, includeOptional bool) []string {
	iter, err := v.Fields(cue.Optional(includeOptional))
	if err != nil {
		return nil
	}
	var names []string
	for iter.Next() {
		sel := iter.Selector()
		if sel.IsDefinition() || sel.PkgPath() != "" {
			continue
		}
		names = append(names, normalizeLabel(sel.String()))
	}
	return names
}

// normalizeLabel strips the trailing optionality marker from a CUE
// selector string so that required and optional variants of the same
// field compare equal (e.g. `"foo"?` becomes `"foo"`).
func normalizeLabel(s string) string {
	return strings.TrimSuffix(s, "?")
}

func contains(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}

func collectStableDefs(schemaDir string) ([]string, error) {
	var stableDefs []string

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("read schema dir: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".cue") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(schemaDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		content := string(data)
		if !strings.Contains(content, `@status("stable")`) {
			continue
		}
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") && strings.Contains(line, ":") {
				def := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
				stableDefs = append(stableDefs, def)
			}
		}
	}
	return stableDefs, nil
}

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
