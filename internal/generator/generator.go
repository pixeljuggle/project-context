package generator

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Rule struct {
	Content bool `json:"content"`
}

type Config struct {
	Ignores []string        `json:"ignores,omitempty"`
	Rules   map[string]Rule `json:"rules,omitempty"`
}

// isIgnored hard-codes .git (and everything under it)
func isIgnored(relPath string, patterns []string) bool {
	if relPath == "." || relPath == "" {
		return false
	}

	relPath = filepath.ToSlash(relPath)

	// HARD-CODED: Always ignore .git directory and everything inside it.
	if after := strings.TrimPrefix(relPath, ".git"); after != relPath {
		if after == "" || strings.HasPrefix(after, "/") {
			return true
		}
	}

	// ... rest of the ignore logic
	base := filepath.Base(relPath)

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") || strings.HasPrefix(pattern, "!") {
			continue
		}

		// Fixed: S1017 – use unconditional TrimPrefix instead of if + HasPrefix + slice
		pattern = strings.TrimPrefix(pattern, "/")

		trimmed := strings.TrimSuffix(pattern, "/")
		isDirPattern := strings.HasSuffix(pattern, "/")

		trimmed = strings.ReplaceAll(trimmed, "**", "*")

		matched := false
		var err error

		if strings.Contains(trimmed, "/") {
			matched, err = filepath.Match(trimmed, relPath)
		} else {
			matched, err = filepath.Match(trimmed, base)
		}
		if err == nil && matched {
			return true
		}

		if isDirPattern {
			dir := trimmed
			if relPath == dir || strings.HasPrefix(relPath, dir+"/") {
				return true
			}
		}
	}
	return false
}

func getContentRule(relPath string, rules map[string]Rule) bool {
	relPath = filepath.ToSlash(relPath)
	for p, rule := range rules {
		p = filepath.ToSlash(strings.TrimSuffix(p, "/"))
		if relPath == p || strings.HasPrefix(relPath, p+"/") {
			return rule.Content
		}
	}
	return true
}

func GenerateTreeAndFiles(root string, ignorePatterns []string, rules map[string]Rule) (tree string, contentFiles []string, err error) {
	var treeBuilder strings.Builder
	var files []string

	getRel := func(p string) string {
		rel, _ := filepath.Rel(root, p)
		return filepath.ToSlash(rel)
	}

	var recurse func(currentDir string, prefix string) error
	recurse = func(currentDir string, prefix string) error {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return err
		}

		var valid []fs.DirEntry
		for _, entry := range entries {
			full := filepath.Join(currentDir, entry.Name())
			rel := getRel(full)
			if isIgnored(rel, ignorePatterns) {
				continue
			}
			valid = append(valid, entry)
		}

		for i, entry := range valid {
			isLast := i == len(valid)-1
			name := entry.Name()
			fullPath := filepath.Join(currentDir, name)
			relPath := getRel(fullPath)

			connector := "├── "
			if isLast {
				connector = "└── "
			}

			treeBuilder.WriteString(prefix + connector + name)
			if entry.IsDir() {
				treeBuilder.WriteString("/\n")
				newPrefix := prefix
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
				if err := recurse(fullPath, newPrefix); err != nil {
					return err
				}
			} else {
				treeBuilder.WriteString("\n")
				if getContentRule(relPath, rules) {
					files = append(files, relPath)
				}
			}
		}
		return nil
	}

	absRoot, _ := filepath.Abs(root)
	baseName := filepath.Base(absRoot)
	if baseName == "." || baseName == string(filepath.Separator) {
		baseName = "project"
	}
	treeBuilder.WriteString(baseName + "/\n")

	if err := recurse(root, ""); err != nil {
		return "", nil, err
	}

	return treeBuilder.String(), files, nil
}
