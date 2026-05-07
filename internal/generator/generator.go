package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Rule struct {
	Content bool `json:"content"`
}

type Config struct {
	Ignores []string        `json:"ignores,omitempty"`
	Rules   map[string]Rule `json:"rules,omitempty"`
}

// BuildMarkdown now safely handles Markdown files containing ``` code blocks
func BuildMarkdown(tree string, contentFiles []string, root string, maxSizeBytes int64) string {
	var md strings.Builder
	md.WriteString("# Project Context\n\n")
	md.WriteString("**Estimated tokens:** ~" + estimateTokens(tree, contentFiles, root, maxSizeBytes) + "\n\n")
	md.WriteString("## Directory Tree\n\n")
	md.WriteString("```\n")
	md.WriteString(tree)
	md.WriteString("```\n\n")
	md.WriteString("## File Contents\n\n")

	for _, relPath := range contentFiles {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		if maxSizeBytes > 0 && int64(len(data)) > maxSizeBytes || isBinary(data) {
			md.WriteString("### " + relPath + "\n\n")
			md.WriteString("_**Note:** File skipped (binary or exceeds --max-size limit)_\n\n")
			continue
		}

		content := string(data)
		ext := filepath.Ext(relPath)
		lang := strings.TrimPrefix(ext, ".")
		if lang == "" {
			lang = "plaintext"
		}

		// === FIX: Use 4 backticks for any Markdown file ===
		fence := "```"
		if lang == "md" || lang == "markdown" || lang == "mdx" {
			fence = "````"
		}

		md.WriteString("### " + relPath + "\n\n")
		md.WriteString(fence + lang + "\n")
		md.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			md.WriteString("\n")
		}
		md.WriteString(fence + "\n\n")
	}
	return md.String()
}

func estimateTokens(tree string, contentFiles []string, root string, maxSizeBytes int64) string {
	total := len(tree)
	for _, relPath := range contentFiles {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		if maxSizeBytes > 0 && int64(len(data)) > maxSizeBytes || isBinary(data) {
			continue
		}
		total += len(data)
	}
	return fmt.Sprintf("%d", total/4+300)
}

func isBinary(data []byte) bool {
	for i := 0; i < len(data) && i < 512; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

// isIgnored with hard-coded common junk folders
func isIgnored(relPath string, patterns []string) bool {
	if relPath == "." || relPath == "" {
		return false
	}

	relPath = filepath.ToSlash(relPath)

	// Hard-coded ignores (S1017 compliant)
	hardCoded := []string{
		".git", "node_modules", "dist", "build", "target",
		"venv", ".venv", ".next", "__pycache__", "coverage",
	}
	for _, d := range hardCoded {
		if after := strings.TrimPrefix(relPath, d); after != relPath {
			if after == "" || strings.HasPrefix(after, "/") {
				return true
			}
		}
	}

	base := filepath.Base(relPath)

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") || strings.HasPrefix(pattern, "!") {
			continue
		}

		pattern = strings.TrimPrefix(pattern, "/")
		trimmed := strings.TrimSuffix(pattern, "/")
		isDirPattern := strings.HasSuffix(pattern, "/")

		trimmed = strings.ReplaceAll(trimmed, "**", "*")

		var matched bool
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
			if relPath == trimmed || strings.HasPrefix(relPath, trimmed+"/") {
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

		sort.Slice(valid, func(i, j int) bool {
			return valid[i].Name() < valid[j].Name()
		})

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
