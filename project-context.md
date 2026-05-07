# Project Context

**Estimated tokens:** ~363

## Directory Tree

```
project-context-cli/
├── .gitignore
├── Makefile
├── README.md
├── cmd/
├── go.mod
└── internal/
    └── generator/
        └── generator.go
```

## File Contents

### .gitignore

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Go build artifacts
project-context
project-context.exe

# Build output
bin/
dist/

# Test coverage
*.out
*.test
*.prof

# Dependency directories (if you ever vendor)
vendor/

# IDE / Editor
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Project specific
project-context.json
project-context.md
```

### Makefile

```plaintext
.PHONY: all build clean install run linux darwin windows

BINARY_NAME := project-context
VERSION     ?= v0.3.0

build:
	go build -ldflags "-s -w -X main.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd/project-context

install:
	go install -ldflags "-s -w -X main.Version=$(VERSION)" ./cmd/project-context

run:
	go run ./cmd/project-context

all: linux darwin windows

linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/project-context
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/project-context

darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/project-context
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/project-context

windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/project-context

clean:
	rm -rf bin/ $(BINARY_NAME) $(BINARY_NAME).exe
```

### README.md

````md
# Project Context CLI

Fast, zero-dependency tool that generates a clean `project-context.md` perfect for LLMs (Claude, Cursor, Aider, Grok, etc.).

### Features

- Beautiful directory tree
- File contents in properly tagged code blocks
- Hard-coded ignores for `.git`, `node_modules`, `dist`, `build`, `target`, `venv`, `.next`, etc.
- Smart skipping of binary files and large files (`--max-size`)
- Token count estimation
- `--stdout` support (pipe to clipboard/LLM)
- Full `.gitignore` + `project-context.json` support
- Colored terminal UX

### Quick Start

```bash
# Normal use
project-context

# Pipe directly to LLM / clipboard
project-context --stdout | pbcopy

# Custom options
project-context --max-size 500 --stdout
```

### New Flags

- `--stdout` → Print to stdout instead of file
- `--max-size int` → Max file size in KB for content (default 1024, 0 = unlimited)
- `--version` → Show version
- All previous flags still work (`-I`, `-config`, etc.)

See `project-context --help` for full details.
````

### go.mod

```mod
module github.com/pixeljuggle/project-context

go 1.26.1
```

### internal/generator/generator.go

```go
package generator

import (
	"fmt"
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

// BuildMarkdown now safely handles Markdown files containing ``` code blocks
func BuildMarkdown(tree string, contentFiles []string, root string, maxSizeBytes int64) string {
	var md strings.Builder
	md.WriteString("# Project Context\n\n")
	md.WriteString("**Estimated tokens:** ~" + estimateTokens(tree+strings.Join(contentFiles, "")) + "\n\n")
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

func estimateTokens(s string) string {
	// Rough but useful estimate (4 chars ≈ 1 token)
	return fmt.Sprintf("%d", len(s)/4+300)
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
```

