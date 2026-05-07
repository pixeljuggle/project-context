# Project Context

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
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build for current platform
build:
	go build -ldflags "-s -w -X main.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd/project-context

# Install globally
install:
	go install -ldflags "-s -w -X main.Version=$(VERSION)" ./cmd/project-context

# Run directly
run:
	go run ./cmd/project-context

# Build for all popular platforms
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

```md
# Project Context CLI

A fast, zero-dependency CLI tool that generates a perfect `project-context.md` file for LLMs, code reviews, or documentation.

It includes:

- Beautiful recursive directory tree
- All file contents in properly tagged code blocks
- Full support for `.gitignore`
- Hard-coded ignore for `.git` (and everything inside it)
- JSON config for custom ignores + per-folder rules
- Extra ignore rules via `-I` flag

## Installation

```bash
# Install latest version globally
go install github.com/pixeljuggle/project-context/cmd/project-context@latest

# Or build from source
git clone https://github.com/pixeljuggle/project-context.git
cd project-context
make install
```

## Quick Start

```bash
# In any project
project-context

# Custom options
project-context -root ./my-app -config my-config.json -I "*.tmp" -I "build/"
```

## Configuration (`project-context.json`)

```json
{
  "ignores": ["node_modules/", "dist/", "*.log"],
  "rules": {
    "docs/": { "content": false },
    "internal/secret.txt": { "content": false }
  }
}
```

## Makefile Commands

```bash
make build          # Build for current OS/arch
make install        # Install to $GOPATH/bin
make run            # Run directly
make all            # Build all popular platforms
make clean
```

See `Makefile` for full cross-compilation targets (Linux, macOS, Windows, arm64, etc.).
```

### go.mod

```mod
module github.com/pixeljuggle/project-context

go 1.26.1
```

### internal/generator/generator.go

```go
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
```

