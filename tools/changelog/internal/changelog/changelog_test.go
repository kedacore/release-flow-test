package changelog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveVersionFromReleaseBranch(t *testing.T) {
	repo := t.TempDir()

	mustRun(t, repo, "git", "init")
	mustRun(t, repo, "git", "checkout", "-b", "release/v2.18")
	mustRun(t, repo, "git", "config", "user.email", "test@example.com")
	mustRun(t, repo, "git", "config", "user.name", "test")
	mustWrite(t, filepath.Join(repo, "README.md"), "seed\n")
	mustRun(t, repo, "git", "add", ".")
	mustRun(t, repo, "git", "commit", "-m", "seed")

	version, err := ResolveVersion(repo, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != "2.18" {
		t.Fatalf("expected version 2.18, got %s", version)
	}
}

func TestLoadEntriesAllowsAnyFilenameWithYamlExtension(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "new-feature.yaml"), "Description: Add foo\nType: feat\nIssues:\n  - \"#123\"\n")

	entries, err := LoadEntries(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].FileName != "new-feature.yaml" {
		t.Fatalf("expected filename new-feature.yaml, got %s", entries[0].FileName)
	}
}

func TestLoadEntriesRequiresIssueOrPR(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "missing-ref.yaml"), "Description: Add foo\nType: feat\n")

	_, err := LoadEntries(dir)
	if err == nil {
		t.Fatal("expected validation error when both Issues and PRs are missing")
	}
	if !strings.Contains(err.Error(), "at least one reference is required") {
		t.Fatalf("expected missing reference error, got %v", err)
	}
}

func TestUpsertVersionSection(t *testing.T) {
	existing := "# Changelog\n\n## History\n\n- [Unreleased](#unreleased)\n- [v1.0.0](#v100)\n\n## Unreleased\n\n### New\n\nNone.\n\n## v1.0.0\n\n### New\n\n- **General**: A\n"
	newSection := "## v1.1.0\n\n### New\n\n- **General**: B\n"

	updated := UpsertVersionSection(existing, "1.1.0", newSection)
	if updated == "" {
		t.Fatal("expected non-empty updated changelog")
	}
	if !strings.Contains(updated, "## v1.0.0") || !strings.Contains(updated, "## v1.1.0") {
		t.Fatalf("expected both versions in output, got:\n%s", updated)
	}
	if !strings.Contains(updated, "- [v1.1.0](#v110)") {
		t.Fatalf("expected v1.1.0 in history, got:\n%s", updated)
	}
}

func TestRenderVersionSectionKEDAStyle(t *testing.T) {
	entries := []Entry{
		{
			Description:  "Add new scaler behavior",
			Type:         "feat",
			Component:    "scaler",
			Subcomponent: "operator",
			Issues:       []string{"#123"},
			FileName:     "a.yaml",
		},
	}

	rendered := RenderVersionSection("1.0.0", entries)
	if !strings.Contains(rendered, "## v1.0.0") {
		t.Fatalf("expected version header, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "### New") {
		t.Fatalf("expected New section, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "- **Scaler(Operator)**: Add new scaler behavior ([#123](https://github.com/kedacore/keda/issues/123))") {
		t.Fatalf("expected KEDA-style bullet with issue link, got:\n%s", rendered)
	}
}

func TestRenderVersionSectionWithPROnly(t *testing.T) {
	entries := []Entry{
		{
			Description: "Improve scaler internals",
			Type:        "improvement",
			PRs:         []string{"456"},
			FileName:    "b.yaml",
		},
	}

	rendered := RenderVersionSection("1.0.1", entries)
	if !strings.Contains(rendered, "([#456](https://github.com/kedacore/keda/pull/456))") {
		t.Fatalf("expected PR link when issue is missing, got:\n%s", rendered)
	}
}

func TestRenderVersionSectionWithIssuesAndPRs(t *testing.T) {
	entries := []Entry{
		{
			Description: "Fix scaler panic",
			Type:        "fix",
			Issues:      []string{"#123"},
			PRs:         []string{"#456"},
			FileName:    "c.yaml",
		},
	}

	rendered := RenderVersionSection("1.0.2", entries)
	if !strings.Contains(rendered, "[#123](https://github.com/kedacore/keda/issues/123), [#456](https://github.com/kedacore/keda/pull/456)") {
		t.Fatalf("expected both issue and PR links, got:\n%s", rendered)
	}
}

func TestRenderEntryTemplateFromSchema(t *testing.T) {
	template, err := renderEntryTemplateFromSchema()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, key := range []string{"Description:", "Type:", "Component:", "Subcomponent:", "Issues:", "PRs:"} {
		if !strings.Contains(template, key) {
			t.Fatalf("expected template to include %s, got:\n%s", key, template)
		}
	}

	if !strings.Contains(template, "# Example: #123") || !strings.Contains(template, "# Example: #456") {
		t.Fatalf("expected template to include examples from schema tags, got:\n%s", template)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s failed: %v", path, err)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command %s %v failed: %v\n%s", name, args, err, string(out))
	}
}
