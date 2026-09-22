package activity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/srz-zumix/go-gh-extension/pkg/ioutil"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

const defaultFilePerm = 0o644

// KindData is the payload written to <kind>.json. Data holds the raw records
// as returned by the GitHub API, so the JSON files keep every field the
// Markdown rendering drops.
type KindData struct {
	Kind  Kind   `json:"kind"`
	Owner string `json:"owner,omitempty"`
	Count int    `json:"count"`
	Data  any    `json:"data"`
}

// WriteOptions configures a Write call.
type WriteOptions struct {
	// Dir is the output directory.
	Dir string
	// Mode selects how much detail is rendered for each kind.
	Mode Mode
	// Kinds is the set of activity kinds to write.
	Kinds []Kind
	// SkipEmpty suppresses the output files of kinds with no collected data.
	SkipEmpty bool
	// Force empties Dir before writing instead of failing when it is not empty.
	Force bool
}

// Write renders res and writes Markdown and JSON files for every kind in
// opts.Kinds under opts.Dir. Root-scoped kinds are written directly under the
// directory; owner-scoped kinds are written under <dir>/<owner>/. A top-level
// summary.md overview and an AGENTS.md guide to the output layout are always
// written. Writing fails when the directory already contains files, unless
// opts.Force is set.
func Write(res *Result, opts WriteOptions) error {
	dir := opts.Dir
	mode := opts.Mode
	kinds := opts.Kinds
	logger.Info("writing activity output", "dir", dir, "mode", mode, "kinds", len(kinds), "skip_empty", opts.SkipEmpty, "force", opts.Force)
	if err := prepareDir(dir, opts.Force); err != nil {
		return err
	}

	for _, kind := range kinds {
		if isRootKind(kind) {
			section := res.Section(kind)
			if opts.SkipEmpty && section.IsEmpty() {
				logger.Debug("skipped empty activity kind", "kind", kind)
				continue
			}
			data := KindData{Kind: kind, Count: section.Count, Data: res.Data(kind)}
			if err := writeKindFiles(dir, section, data, mode); err != nil {
				return err
			}
			continue
		}

		for _, owner := range res.OwnerLogins() {
			oa := res.Owners[owner]
			section := oa.Section(kind)
			if opts.SkipEmpty && section.IsEmpty() {
				logger.Debug("skipped empty activity kind", "kind", kind, "owner", owner)
				continue
			}
			ownerDir := filepath.Join(dir, owner)
			if err := os.MkdirAll(ownerDir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory '%s': %w", ownerDir, err)
			}
			data := KindData{Kind: kind, Owner: owner, Count: section.Count, Data: oa.Data(kind)}
			if err := writeKindFiles(ownerDir, section, data, mode); err != nil {
				return err
			}
		}
	}

	overview := RenderSummaryOverview(res, kinds)
	summaryPath := filepath.Join(dir, "summary.md")
	if err := ioutil.WriteFileAtomic(summaryPath, []byte(overview), defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write summary file in '%s': %w", dir, err)
	}

	guide := RenderAgentsGuide(res, mode, kinds, opts.SkipEmpty)
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := ioutil.WriteFileAtomic(agentsPath, []byte(guide), defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write agents guide in '%s': %w", dir, err)
	}

	logger.Info("wrote activity output", "dir", dir, "summary", summaryPath, "agents", agentsPath, "warnings", len(res.Warnings))
	return nil
}

// CheckOutputDir reports whether dir can be used as an output directory,
// without modifying it. A directory that already holds files is an error
// unless force is set.
func CheckOutputDir(dir string, force bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read output directory '%s': %w", dir, err)
	}
	if len(entries) == 0 || force {
		return nil
	}
	return fmt.Errorf("output directory '%s' is not empty: use --force to empty it before writing", dir)
}

// prepareDir creates dir if needed and makes sure it is empty. A non-empty
// directory is an error unless force is set, in which case its contents are
// removed.
func prepareDir(dir string, force bool) error {
	if err := CheckOutputDir(dir, force); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", dir, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read output directory '%s': %w", dir, err)
	}
	if len(entries) == 0 {
		return nil
	}

	logger.Info("emptying output directory", "dir", dir, "entries", len(entries))
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove '%s' from output directory: %w", path, err)
		}
	}
	return nil
}

func writeKindFiles(dir string, section KindSection, data KindData, mode Mode) error {
	kind := section.Kind
	mdPath := filepath.Join(dir, string(kind)+".md")
	md := RenderMarkdown(section, mode)
	if err := ioutil.WriteFileAtomic(mdPath, []byte(md), defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write '%s': %w", mdPath, err)
	}

	jsonPath := filepath.Join(dir, string(kind)+".json")
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal '%s' as JSON: %w", kind, err)
	}
	if err := ioutil.WriteFileAtomic(jsonPath, raw, defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write '%s': %w", jsonPath, err)
	}

	logger.Debug("wrote activity kind files", "kind", kind, "count", section.Count, "markdown", mdPath, "json", jsonPath)
	return nil
}
