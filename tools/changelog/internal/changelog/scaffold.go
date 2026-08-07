package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const DefaultEntriesDir = ".changelog"

func ResolveEntriesDir(repoRoot, entriesDir string) string {
	entriesDir = strings.TrimSpace(entriesDir)
	if entriesDir == "" {
		entriesDir = DefaultEntriesDir
	}
	if filepath.IsAbs(entriesDir) {
		return entriesDir
	}
	if strings.TrimSpace(repoRoot) == "" {
		return entriesDir
	}
	return filepath.Join(repoRoot, entriesDir)
}

func CreateEntryFile(entriesDir, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("entry name is required")
	}
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		name += ".yaml"
	}

	path := filepath.Join(entriesDir, name)
	if err := EnsureParentDir(path); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("entry already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return "", err
	}

	content, err := renderEntryTemplateFromSchema()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func renderEntryTemplateFromSchema() (string, error) {
	t := reflect.TypeOf(Entry{})
	lines := make([]string, 0, t.NumField()*5)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlName := yamlFieldName(field)
		if yamlName == "" || yamlName == "-" {
			continue
		}

		doc := strings.TrimSpace(field.Tag.Get("doc"))
		if doc != "" {
			lines = append(lines, "# "+doc)
		}

		example := strings.TrimSpace(field.Tag.Get("example"))
		if example != "" {
			lines = append(lines, "# Example: "+example)
		}

		switch field.Type.Kind() {
		case reflect.String:
			lines = append(lines, fmt.Sprintf("%s: \"\"", yamlName))
		case reflect.Slice:
			if field.Type.Elem().Kind() != reflect.String {
				return "", fmt.Errorf("unsupported slice type for field %s", field.Name)
			}
			lines = append(lines, fmt.Sprintf("%s:", yamlName))
			if example != "" {
				lines = append(lines, fmt.Sprintf("  - \"%s\"", example))
			} else {
				lines = append(lines, "  - \"\"")
			}
		default:
			return "", fmt.Errorf("unsupported field type for %s", field.Name)
		}

		lines = append(lines, "")
	}

	return strings.Join(lines, "\n"), nil
}

func yamlFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return field.Name
	}
	parts := strings.Split(tag, ",")
	return strings.TrimSpace(parts[0])
}
