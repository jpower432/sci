// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// breakingCheckOpts configures a breaking-change check run.
type breakingCheckOpts struct {
	schemaDir     string // CUE package directory used to generate the candidate
	basePath      string // Baseline OpenAPI file to diff against (required)
	candidatePath string // Local candidate OpenAPI file; empty means generate from schemaDir
	allowPath     string // Allowlist file of oasdiff check IDs; empty means no allowlist
}

// filterAllowed drops changes matched by the allowlist. An entry keyed
// "<id> <path>" drops only that change on that schema path.
func filterAllowed(changes []Change, allowed map[string]bool) []Change {
	var out []Change
	for _, c := range changes {
		if allowed[c.ID+" "+c.Path] {
			continue
		}
		out = append(out, c)
	}
	return out
}

// loadAllowlist reads an allowlist file of oasdiff exceptions, one per line.
// Each entry is a path-scoped "<id> <path>" pair (see filterAllowed). Blank lines
// and lines beginning with '#' are ignored. An empty
// path yields an empty (non-nil) map and a nil error.
func loadAllowlist(path string) (allowed map[string]bool, err error) {
	allowed = make(map[string]bool)
	if path == "" {
		return allowed, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening allowlist %q: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing allowlist %q: %w", path, cerr)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		allowed[line] = true
	}
	if serr := scanner.Err(); serr != nil {
		return nil, fmt.Errorf("reading allowlist %q: %w", path, serr)
	}
	return allowed, err
}

// runBreakingCheck diffs a candidate OpenAPI projection against the baseline at
// opts.basePath, filters allowlisted changes, prints each remaining breaking
// change, and returns exit code 1 when any remain (0 otherwise). The baseline is
// supplied by the caller (fetched from releases in CI/Makefile via gh); an empty
// basePath is an error.
func runBreakingCheck(opts breakingCheckOpts) (exitCode int, err error) {
	if opts.basePath == "" {
		return 0, fmt.Errorf("--base is required (path to the baseline OpenAPI file)")
	}

	candidatePath := opts.candidatePath
	if candidatePath == "" {
		tmpDir, mkErr := os.MkdirTemp("", "gemara-breaking-candidate-")
		if mkErr != nil {
			return 0, fmt.Errorf("creating candidate temp dir: %w", mkErr)
		}
		defer func() {
			if rerr := os.RemoveAll(tmpDir); rerr != nil && err == nil {
				err = fmt.Errorf("removing candidate temp dir: %w", rerr)
			}
		}()
		candidatePath = filepath.Join(tmpDir, "openapi.yaml")
		if cerr := convertCUEToOpenAPI(opts.schemaDir, candidatePath, ConvertOpts{}); cerr != nil {
			return 0, fmt.Errorf("generating candidate OpenAPI: %w", cerr)
		}
	}

	basePath := opts.basePath

	base, err := loadWrapped(basePath)
	if err != nil {
		return 0, fmt.Errorf("loading baseline %q: %w", basePath, err)
	}
	rev, err := loadWrapped(candidatePath)
	if err != nil {
		return 0, fmt.Errorf("loading candidate %q: %w", candidatePath, err)
	}

	changes, err := breakingChanges(base, rev)
	if err != nil {
		return 0, fmt.Errorf("computing breaking changes: %w", err)
	}

	allowed, err := loadAllowlist(opts.allowPath)
	if err != nil {
		return 0, err
	}

	remaining := filterAllowed(changes, allowed)
	for _, c := range remaining {
		fmt.Printf("ERR[%s] %s: %s\n", c.ID, c.Path, c.Text)
	}
	if len(remaining) > 0 {
		return 1, nil
	}
	return 0, nil
}

// newBreakingCheckCmd builds the breaking-check cobra subcommand.
func newBreakingCheckCmd() *cobra.Command {
	opts := breakingCheckOpts{}
	cmd := &cobra.Command{
		Use:   "breaking-check",
		Short: "Detect backward-incompatible schema changes against the v1 baseline",
		Long: `Generate an OpenAPI projection of the current CUE schema and diff it against
a baseline OpenAPI file using oasdiff. The baseline is supplied via --base
(fetched from the latest v1 release by the Makefile/CI). Backward-incompatible
(ERR-level) changes fail the check unless their oasdiff check ID is listed in the
allowlist.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := runBreakingCheck(opts)
			if err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("breaking changes detected")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.schemaDir, "schema", "../..", "Path to the CUE package directory")
	cmd.Flags().StringVar(&opts.basePath, "base", "", "Path to the baseline OpenAPI file to diff against (required)")
	cmd.Flags().StringVar(&opts.allowPath, "allow", "", "Path to an allowlist file of path-scoped oasdiff exceptions (\"<id> <path>\", one per line)")
	return cmd
}
