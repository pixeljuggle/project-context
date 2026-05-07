# Project Context

## Directory Tree

```
project-context-cli/
├── .github/
│   └── workflows/
│       └── release.yml
├── .gitignore
├── .goreleaser.yaml
├── Makefile
├── README.md
├── cmd/
│   └── project-context/
│       └── main.go
├── docs/
├── go.mod
└── internal/
    └── generator/
        └── generator.go
```

## File Contents

### .github/workflows/release.yml

```yml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: stable
          cache: true

      - uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### .gitignore

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib


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

```

### .goreleaser.yaml

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    binary: project-context
    ldflags:
      - -s -w -X main.Version={{.Version}}

archives:
  - formats: ["tar.gz"]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ with .Arm }}v{{ . }}{{ end }}"
    format_overrides:
      - goos: windows
        formats: ["zip"]
    files:
      - LICENSE*
      - README.md
      - CHANGELOG.md

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: asc
  use: github
  groups:
    - title: Features
      regexp: "^feat:"
      order: 100
    - title: Bug Fixes
      regexp: "^fix:"
      order: 200
    - title: Others
      regexp: "^(chore|docs|refactor|test|style):"
      order: 300

release:
  draft: false
  prerelease: auto
```

### Makefile

```plaintext
GOBIN := $(shell go env GOPATH)/bin

.PHONY: all build clean install run linux darwin windows release release-dry-run lint

BINARY_NAME := project-context
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build for current platform → bin/
build:
	mkdir -p bin
	go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME) ./cmd/project-context

# Install globally
install:
	go install -ldflags "-s -w -X main.Version=$(VERSION)" ./cmd/project-context

# Run directly
run:
	go run ./cmd/project-context -I "docs/project-context.md" -output "docs/project-context.md" 

# Build for all platforms → bin/
all: linux darwin windows

linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/project-context
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/project-context

darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/project-context
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/project-context

windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/project-context


# === Release targets (auto-install GoReleaser) ===
release:
	@$(call install-tool,goreleaser,github.com/goreleaser/goreleaser/v2@latest)
	$(GOBIN)/goreleaser release --snapshot --clean

release-dry-run:
	@$(call install-tool,goreleaser,github.com/goreleaser/goreleaser/v2@latest)
	$(GOBIN)/goreleaser check

# === Lint target (auto-install staticcheck) ===
lint:
	@echo "Running gofmt..."
	@test -z "$$(gofmt -l .)" || (echo "❌ gofmt issues found:" && gofmt -l . && exit 1)
	@echo "Running go vet..."
	go vet ./...
	@echo "Running staticcheck..."
	@$(call install-tool,staticcheck,honnef.co/go/tools/cmd/staticcheck@latest)
	$(GOBIN)/staticcheck ./...
	@echo "✅ All lint checks passed!"

# Helper to auto-install Go tools
define install-tool
	@if ! test -x $(GOBIN)/$(1) && ! command -v $(1) >/dev/null 2>&1; then \
		echo "🔧 Installing $(1)..."; \
		go install $(2); \
	fi
endef

clean:
	rm -rf bin/ dist/ $(BINARY_NAME) $(BINARY_NAME).exe

# Quick test after build
test-build:
	./bin/$(BINARY_NAME) --version
```

### README.md

```md
# Project Context CLI

## A fast, zero-dependency CLI tool that generates a perfect `project-context.md` for LLMs, code reviews, or documentation.

## Features

- Recursive directory tree (Git-style)
- All file contents in properly syntax-highlighted code blocks
- **Safe Markdown handling** — uses 4-backtick fences for `.md` files so nested code blocks never break the output
- Hard-coded ignores for common junk: `.git/`, `node_modules/`, `dist/`, `build/`, `target/`, `venv/`, `.venv/`, `.next/`, `__pycache__`, etc.
- Full `.gitignore` support + custom `project-context.json` config
- Smart skipping of binary files and large files (`--max-size`)
- Token estimation (rough but useful for LLM context limits)
- `--stdout` mode — perfect for piping directly to clipboard or AI tools
- Zero external dependencies (pure Go stdlib)

## Installation

### From GitHub Releases (recommended)

Download the latest binary for your OS from the [Releases page](https://github.com/pixeljuggle/project-context/releases).

### With Go

```bash
go install github.com/pixeljuggle/project-context/cmd/project-context@latest
```

## Usage

```bash
project-context --version
project-context -I "node_modules/" -I "dist/"
project-context -config my-config.json
```

## Automatic Releases

This project uses **GoReleaser** + GitHub Actions.  
Just run:

```bash
git tag v1.0.0
git push origin v1.0.0
```

A new release with binaries for Linux, macOS, and Windows will be created automatically.

See `.goreleaser.yaml` and `.github/workflows/release.yml` for details.

## Quick Start

```bash
# 1. Go to any project
cd ~/my-awesome-app

# 2. Generate the context file
project-context

# 3. (Optional) Pipe directly to clipboard for instant LLM use
project-context --stdout | pbcopy
```

### Realistic everyday examples

**Example 1: Full context for a Next.js / TypeScript project**

```bash
project-context --max-size 500 --stdout | pbcopy
```

**Example 2: Ignore extra patterns on the fly**

```bash
project-context -I "*.test.ts" -I "e2e/" -I "coverage/"
```

**Example 3: Use a custom config and output to a different name**

```bash
project-context -config my-context-config.json -output context-for-claude.md
```

**Example 4: Skip .gitignore and only use your own rules**

```bash
project-context --no-gitignore -I "node_modules/" -I "dist/"
```

## Configuration (`project-context.json`)

Create this file in your project root for project-specific rules:

```json
{
  "ignores": ["*.log", "*.tmp", "coverage/", "storybook-static/"],
  "rules": {
    "docs/": { "content": false },
    "public/": { "content": false },
    "src/assets/": { "content": false },
    "internal/secret-config.ts": { "content": false }
  }
}
```

- `ignores`: `.gitignore`-style patterns (applied in addition to `.gitignore`)
- `rules`: Per-folder or per-file rules (`content: false` = show in tree but skip content)

## Command-Line Flags

| Flag             | Default                | Description                             |
| ---------------- | ---------------------- | --------------------------------------- |
| `--stdout`       | false                  | Print to stdout instead of writing file |
| `--max-size`     | 1024                   | Max file size in KB (0 = unlimited)     |
| `--root`         | `.`                    | Project root directory                  |
| `--config`       | `project-context.json` | Path to config file                     |
| `--output`       | `project-context.md`   | Output markdown filename                |
| `--no-gitignore` | false                  | Skip loading `.gitignore`               |
| `-I`             | (repeatable)           | Additional ignore pattern               |
| `--version`      | -                      | Show version and exit                   |
| `--help`         | -                      | Show help                               |

## Makefile Commands

```bash
make build             # Build for current OS/arch → ./bin/project-context
make install           # Install globally via go install
make run               # Run directly with go run
make all               # Build all platforms (Linux, macOS, Windows + amd64/arm64) → ./bin/
make release           # Full local release test (creates ./dist/ with .tar.gz + .zip)
make release-dry-run   # Only validate .goreleaser.yaml config (no build)
make lint              # Run gofmt + go vet + staticcheck (recommended before committing)
make clean             # Remove bin/ and dist/
make test-build        # Quick version check after build
```

**Note about releases & linting:**

- `make release` uses GoReleaser in snapshot mode for local testing.
- `make release-dry-run` validates the release config only.
- `make lint` is the recommended pre-commit check (it will fail CI-style if formatting or issues are found).
```

### cmd/project-context/main.go

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pixeljuggle/project-context/internal/generator"
)

var Version = "dev" // will be overridden by ldflags during release

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var ignoreFlags stringSlice

	rootDir := flag.String("root", ".", "Project root directory")
	configFile := flag.String("config", "project-context.json", "Path to JSON config (optional)")
	outputFile := flag.String("output", "project-context.md", "Output Markdown file")
	noGitignore := flag.Bool("no-gitignore", false, "Skip loading .gitignore")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Var(&ignoreFlags, "I", "Ignore pattern (repeatable, .gitignore-style)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generates project-context.md with directory tree + file contents.\n")
		fmt.Fprintf(os.Stderr, ".git is hard-coded ignored.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample config (project-context.json):\n")
		fmt.Fprintf(os.Stderr, `{
  "ignores": ["node_modules/", "dist/"],
  "rules": {
    "docs/": {"content": false}
  }
}`)
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("project-context version %s\n", Version)
		os.Exit(0)
	}

	root, err := filepath.Abs(*rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving root: %v\n", err)
		os.Exit(1)
	}

	// Load config
	config := generator.Config{Rules: make(map[string]generator.Rule)}
	configPath := *configFile
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
	}
	if data, err := os.ReadFile(configPath); err == nil {
		if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid config JSON %s: %v\n", configPath, jsonErr)
		}
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: could not read config %s: %v\n", configPath, err)
	}

	// Build ignore list
	var ignorePatterns []string
	if !*noGitignore {
		gitPath := filepath.Join(root, ".gitignore")
		if data, err := os.ReadFile(gitPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					ignorePatterns = append(ignorePatterns, line)
				}
			}
		}
	}
	ignorePatterns = append(ignorePatterns, config.Ignores...)
	ignorePatterns = append(ignorePatterns, ignoreFlags...)

	// Generate
	treeStr, contentFiles, err := generator.GenerateTreeAndFiles(root, ignorePatterns, config.Rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Build Markdown (unchanged)
	var md strings.Builder
	md.WriteString("# Project Context\n\n")
	md.WriteString("## Directory Tree\n\n")
	md.WriteString("```\n")
	md.WriteString(treeStr)
	md.WriteString("```\n\n")
	md.WriteString("## File Contents\n\n")

	for _, relPath := range contentFiles {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", relPath, err)
			continue
		}

		content := string(data)
		ext := filepath.Ext(relPath)
		lang := strings.TrimPrefix(ext, ".")
		if lang == "" {
			lang = "plaintext"
		}

		md.WriteString("### " + relPath + "\n\n")
		md.WriteString("```" + lang + "\n")
		md.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			md.WriteString("\n")
		}
		md.WriteString("```\n\n")
	}

	outPath := *outputFile
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(root, outPath)
	}
	if err := os.WriteFile(outPath, []byte(md.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created %s\n", outPath)
	fmt.Printf("   • Tree + %d file contents included\n", len(contentFiles))
	fmt.Printf("   • .git is always ignored (hard-coded)\n")
}
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

