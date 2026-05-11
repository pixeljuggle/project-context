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
	Ignores       []string        `json:"ignores,omitempty"`
	Rules         map[string]Rule `json:"rules,omitempty"`
	MaxSizeKB     int             `json:"maxSizeKB,omitempty"`
	TruncateLines int             `json:"truncateLines,omitempty"`
	Output        string          `json:"output,omitempty"`
}

// matchesPattern extracts the core matching logic (used by ignore + negation)
func matchesPattern(pattern, relPath string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return false
	}
	pattern = strings.TrimPrefix(pattern, "/")
	trimmed := strings.TrimSuffix(pattern, "/")
	isDirPattern := strings.HasSuffix(pattern, "/")
	trimmed = strings.ReplaceAll(trimmed, "**", "*")

	base := filepath.Base(relPath)
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
	return false
}

// isIgnored now fully supports .gitignore-style negation (!) with "last match wins"
func isIgnored(relPath string, patterns []string) bool {
	if relPath == "." || relPath == "" {
		return false
	}

	relPath = filepath.ToSlash(relPath)

	// Hard-coded ignores are applied first (they cannot be easily negated)
	hardCoded := []string{
		".git", "node_modules", "dist", "build", "target",
		"venv", ".venv", ".next", "__pycache__", "coverage",
	}
	ignored := false
	for _, d := range hardCoded {
		if after := strings.TrimPrefix(relPath, d); after != relPath {
			if after == "" || strings.HasPrefix(after, "/") {
				ignored = true
				break
			}
		}
	}

	// Process all patterns in order — last matching rule wins
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		isNegation := strings.HasPrefix(pattern, "!")
		if isNegation {
			pattern = strings.TrimPrefix(pattern, "!")
		}

		if matchesPattern(pattern, relPath) {
			if isNegation {
				ignored = false
			} else {
				ignored = true
			}
		}
	}
	return ignored
}

func getContentRule(relPath string, rules map[string]Rule) bool {
	relPath = filepath.ToSlash(relPath)
	baseName := filepath.Base(relPath)

	// 1. Process specific rules first (ignoring the wildcard)
	for p, rule := range rules {
		if p == "*" {
			continue
		}
		p = filepath.ToSlash(strings.TrimSuffix(p, "/"))

		// Opt-in recursive check (e.g., "**/package.json")
		if strings.HasPrefix(p, "**/") {
			targetName := strings.TrimPrefix(p, "**/")
			if baseName == targetName {
				return rule.Content
			}
			continue
		}

		// Default strict check: exact relative path or directory prefix
		if relPath == p || strings.HasPrefix(relPath, p+"/") {
			return rule.Content
		}
	}

	// 2. Fallback to wildcard if defined
	if rule, exists := rules["*"]; exists {
		return rule.Content
	}

	return true
}

func matchesAnyExt(relPath string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(relPath))
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		if ext == e {
			return true
		}
	}
	return false
}

func GenerateTreeAndFiles(root string, ignorePatterns []string, rules map[string]Rule, includeExts []string) (tree string, contentFiles []string, err error) {
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

		// === NEW: Alphabetical sorting (predictable & clean tree) ===
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
				if getContentRule(relPath, rules) && matchesAnyExt(relPath, includeExts) {
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

// estimateTokens now actually counts real file content (much more accurate)
func estimateTokens(tree string, contentFiles []string, root string, maxSizeBytes int64, truncateLines int) string {
	total := len(tree)
	for _, relPath := range contentFiles {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		size := len(data)
		if maxSizeBytes > 0 && int64(size) > maxSizeBytes || isBinary(data) {
			if truncateLines == 0 {
				continue
			}
			// still count full size for rough estimate (conservative)
		}
		total += size
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

// BuildMarkdown now supports truncate, verbose, accurate tokens, and safe Markdown
func BuildMarkdown(tree string, contentFiles []string, root string, maxSizeBytes int64, truncateLines int, verbose bool) string {
	var md strings.Builder
	md.WriteString("# Project Context\n\n")
	md.WriteString("**Estimated tokens:** ~" + estimateTokens(tree, contentFiles, root, maxSizeBytes, truncateLines) + "\n\n")
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

		content := string(data)
		skipped := false
		truncated := false

		if maxSizeBytes > 0 && int64(len(data)) > maxSizeBytes || isBinary(data) {
			if truncateLines > 0 {
				lines := strings.Split(content, "\n")
				if len(lines) > truncateLines {
					content = strings.Join(lines[:truncateLines], "\n") + fmt.Sprintf("\n... (truncated to %d lines)", truncateLines)
					truncated = true
				}
			} else {
				skipped = true
			}
		}

		if skipped {
			if verbose {
				fmt.Fprintf(os.Stderr, "  ⏭️  Skipped: %s (binary or exceeds --max-size)\n", relPath)
			}
			md.WriteString("### " + relPath + "\n\n")
			md.WriteString("_**Note:** File skipped (binary or exceeds --max-size limit)_\n\n")
			continue
		}

		if truncated && verbose {
			fmt.Fprintf(os.Stderr, "  ✂️  Truncated: %s (first %d lines)\n", relPath, truncateLines)
		}

		ext := filepath.Ext(relPath)
		lang := strings.TrimPrefix(ext, ".")
		if lang == "" {
			lang = "plaintext"
		}

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
