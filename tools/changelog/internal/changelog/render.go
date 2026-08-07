package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var sectionTitleCaser = cases.Title(language.Und)

var categoryOrder = []string{
	"New",
	"Improvements",
	"Fixes",
	"Deprecations",
	"Breaking Changes",
	"Other",
}

const kedaIssueBaseURL = "https://github.com/kedacore/keda/issues/"
const kedaPullBaseURL = "https://github.com/kedacore/keda/pull/"

func RenderVersionSection(version string, entries []Entry) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## v%s\n\n", version))

	groups := map[string][]Entry{}
	for _, e := range entries {
		t := normalizeCategory(e.Type)
		groups[t] = append(groups[t], e)
	}

	for _, t := range categoryOrder {
		b.WriteString(fmt.Sprintf("### %s\n\n", t))
		if len(groups[t]) == 0 {
			b.WriteString("None.\n\n")
			continue
		}

		sort.SliceStable(groups[t], func(i, j int) bool {
			return groups[t][i].FileName < groups[t][j].FileName
		})

		for _, e := range groups[t] {
			b.WriteString(renderEntryLine(e))
		}
		b.WriteString("\n")
	}

	return strings.TrimSpace(b.String()) + "\n"
}

func UpsertVersionSection(existing, version, newSection string) string {
	existing = strings.TrimSpace(existing)
	newSection = strings.TrimSpace(newSection)
	if existing == "" {
		existing = renderBaseChangelog()
	}

	existing = ensureHistoryEntry(existing, version)
	existing = upsertSection(existing, "## v"+version, newSection)

	return strings.TrimSpace(existing) + "\n"
}

func renderBaseChangelog() string {
	return strings.TrimSpace(`# Changelog

## Deprecations

To learn more about active deprecations, we recommend checking
[GitHub Discussions](https://github.com/kedacore/keda/discussions/categories/deprecations).

## History

- [Unreleased](#unreleased)

## Unreleased

### New

None.

### Improvements

None.

### Fixes

None.

### Deprecations

None.

### Breaking Changes

None.

### Other

None.`) + "\n"
}

func ensureHistoryEntry(existing, version string) string {
	historyHeader := "## History"
	historyEntry := fmt.Sprintf("- [v%s](#v%s)", version, strings.ReplaceAll(version, ".", ""))
	if strings.Contains(existing, historyEntry) {
		return existing
	}

	historyIdx := strings.Index(existing, historyHeader)
	if historyIdx == -1 {
		return strings.TrimSpace(existing) + "\n\n" + historyHeader + "\n\n- [Unreleased](#unreleased)\n" + historyEntry + "\n"
	}

	sectionStart := historyIdx + len(historyHeader)
	nextSectionRel := strings.Index(existing[sectionStart:], "\n## ")
	sectionEnd := len(existing)
	if nextSectionRel != -1 {
		sectionEnd = sectionStart + nextSectionRel
	}

	historyBody := strings.TrimSpace(existing[sectionStart:sectionEnd])
	if strings.Contains(historyBody, "- [Unreleased](#unreleased)") {
		historyBody = strings.Replace(historyBody, "- [Unreleased](#unreleased)", "- [Unreleased](#unreleased)\n"+historyEntry, 1)
	} else if historyBody == "" {
		historyBody = "- [Unreleased](#unreleased)\n" + historyEntry
	} else {
		historyBody = historyEntry + "\n" + historyBody
	}

	prefix := strings.TrimRight(existing[:sectionStart], "\n")
	suffix := ""
	if sectionEnd < len(existing) {
		suffix = strings.TrimLeft(existing[sectionEnd:], "\n")
	}
	if suffix == "" {
		return prefix + "\n\n" + historyBody
	}
	return prefix + "\n\n" + historyBody + "\n\n" + suffix
}

func upsertSection(existing, header, section string) string {
	idx := strings.Index(existing, header)
	if idx == -1 {
		return strings.TrimSpace(existing) + "\n\n" + strings.TrimSpace(section)
	}

	nextIdxRel := strings.Index(existing[idx+len(header):], "\n## v")
	if nextIdxRel == -1 {
		return strings.TrimRight(existing[:idx], "\n") + "\n\n" + strings.TrimSpace(section)
	}

	nextIdx := idx + len(header) + nextIdxRel + 1
	prefix := strings.TrimRight(existing[:idx], "\n")
	suffix := strings.TrimLeft(existing[nextIdx:], "\n")
	if suffix == "" {
		return prefix + "\n\n" + strings.TrimSpace(section)
	}
	return prefix + "\n\n" + strings.TrimSpace(section) + "\n\n" + suffix
}

func normalizeCategory(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "new", "feat", "feature":
		return "New"
	case "improvement", "improvements", "enhancement", "enhance", "chore":
		return "Improvements"
	case "fix", "bug", "bugs", "bugfix", "hotfix":
		return "Fixes"
	case "deprecation", "deprecations", "deprecated":
		return "Deprecations"
	case "breaking", "breaking-change", "breaking changes":
		return "Breaking Changes"
	default:
		return "Other"
	}
}

func renderEntryLine(entry Entry) string {
	component := "General"
	if strings.TrimSpace(entry.Component) != "" {
		component = sectionTitleCaser.String(strings.TrimSpace(entry.Component))
	}
	if strings.TrimSpace(entry.Subcomponent) != "" {
		component = fmt.Sprintf("%s(%s)", component, sectionTitleCaser.String(strings.TrimSpace(entry.Subcomponent)))
	}

	text := strings.TrimSpace(entry.Description)

	refs := formatRefs(entry.Issues, entry.PRs)
	if refs != "" {
		return fmt.Sprintf("- **%s**: %s (%s)\n", component, text, refs)
	}
	return fmt.Sprintf("- **%s**: %s\n", component, text)
}

func formatRefs(issues, prs []string) string {
	if len(issues) == 0 && len(prs) == 0 {
		return ""
	}

	refs := make([]string, 0, len(issues)+len(prs))
	for _, issue := range issues {
		ref, ok := normalizeRef(issue, kedaIssueBaseURL)
		if !ok {
			continue
		}
		refs = append(refs, ref)
	}
	for _, pr := range prs {
		ref, ok := normalizeRef(pr, kedaPullBaseURL)
		if !ok {
			continue
		}
		refs = append(refs, ref)
	}

	if len(refs) == 0 {
		return ""
	}
	return strings.Join(refs, ", ")
}

func normalizeRef(raw, defaultBaseURL string) (string, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", false
	}

	if strings.HasPrefix(v, "#") {
		n := strings.TrimPrefix(v, "#")
		if _, err := strconv.Atoi(n); err == nil {
			return fmt.Sprintf("[#%s](%s%s)", n, defaultBaseURL, n), true
		}
	}

	if _, err := strconv.Atoi(v); err == nil {
		return fmt.Sprintf("[#%s](%s%s)", v, defaultBaseURL, v), true
	}

	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		parts := strings.Split(strings.TrimRight(v, "/"), "/")
		if len(parts) > 0 {
			last := parts[len(parts)-1]
			if _, err := strconv.Atoi(last); err == nil {
				return fmt.Sprintf("[#%s](%s)", last, v), true
			}
		}
		return v, true
	}

	return v, true
}

func EnsureParentDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}

func ParseAllowedTypes(input string) map[string]struct{} {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	allowed := make(map[string]struct{})
	for _, item := range strings.Split(input, ",") {
		item = strings.ToLower(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		allowed[item] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil
	}
	return allowed
}

func ValidateEntryTypes(entries []Entry, allowed map[string]struct{}) error {
	if len(allowed) == 0 {
		return nil
	}

	for _, entry := range entries {
		if _, ok := allowed[strings.ToLower(entry.Type)]; !ok {
			return fmt.Errorf("entry %s has unsupported type %q", entry.FileName, entry.Type)
		}
	}
	return nil
}
