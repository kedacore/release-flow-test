package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Description  string   `yaml:"Description" doc:"One-sentence description shown in the changelog entry." example:"Add support for Foo scaler."`
	Type         string   `yaml:"Type" doc:"Change category used for grouping in the changelog." example:"feat"`
	Component    string   `yaml:"Component,omitempty" doc:"Main area/component for the entry." example:"scaler"`
	Subcomponent string   `yaml:"Subcomponent,omitempty" doc:"Optional subcomponent shown in parenthesis next to Component." example:"operator"`
	Issues       []string `yaml:"Issues,omitempty" doc:"Issue references. At least one reference in Issues or PRs is required." example:"#123"`
	PRs          []string `yaml:"PRs,omitempty" doc:"Pull request references. Use this when there is no issue. At least one reference in Issues or PRs is required." example:"#456"`

	FileName string `yaml:"-"`
}

func LoadEntries(entriesDir string) ([]Entry, error) {
	files, err := os.ReadDir(entriesDir)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".yaml") && !strings.HasSuffix(f.Name(), ".yml") {
			continue
		}

		path := filepath.Join(entriesDir, f.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var entry Entry
		if err := yaml.Unmarshal(raw, &entry); err != nil {
			return nil, fmt.Errorf("file %s: invalid yaml: %w", f.Name(), err)
		}

		entry.Description = strings.TrimSpace(entry.Description)
		entry.Type = strings.TrimSpace(entry.Type)
		entry.Component = strings.TrimSpace(entry.Component)
		entry.Subcomponent = strings.TrimSpace(entry.Subcomponent)
		entry.FileName = f.Name()

		if entry.Description == "" || entry.Type == "" {
			return nil, fmt.Errorf("file %s: Description and Type are required", f.Name())
		}
		if len(entry.Issues) == 0 && len(entry.PRs) == 0 {
			return nil, fmt.Errorf("file %s: at least one reference is required in Issues or PRs", f.Name())
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].FileName < entries[j].FileName })
	if len(entries) == 0 {
		return nil, fmt.Errorf("no yaml entries found in %s", entriesDir)
	}
	return entries, nil
}
