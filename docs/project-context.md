# Project Context

**Estimated tokens:** ~7081

## Directory Tree

```
project-context-cli/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── .gitignore
├── .goreleaser.yaml
├── LICENSE
├── Makefile
├── README.md
├── cmd/
│   └── project-context/
│       └── main.go
├── docs/
├── go.mod
└── internal/
    └── generator/
        ├── generator.go
        └── generator_test.go
```

## File Contents

### .github/workflows/ci.yml

```yml
name: CI

on:
  push:
    branches: [main, master]
  pull_request:

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable
          cache: true

      - name: Lint
        run: make lint

      - name: Test
        run: go test ./... -race -count=1
```

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
          GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
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
    main: ./cmd/project-context
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

# Modern Homebrew Cask configuration (no deprecation warning)
homebrew_casks:
  - repository:
      owner: pixeljuggle
      name: homebrew-project-context
    name: project-context
    homepage: https://github.com/pixeljuggle/project-context
    description: Zero-dependency CLI that generates a perfect project-context.md for LLMs, code reviews, or documentation.
    license: MIT
    binaries:
      - project-context
```

### LICENSE

```plaintext
MIT License

Copyright (c) 2026 alex

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### Makefile

```plaintext
GOBIN := $(shell go env GOPATH)/bin

.PHONY: all build clean install run test linux darwin windows release release-dry-run lint

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

test:
	go test ./... -race -count=1 -v

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
	@echo "All lint checks passed!"

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

````md
# project-context

**Zero-dependency CLI that generates a perfect `project-context.md` for LLMs, code reviews, or documentation.**

It recursively builds a clean Git-style directory tree and includes the contents of relevant source files — with safe Markdown handling, full `.gitignore` support (including `!` negation), binary skipping, size limits, truncation, and extension filtering.

See a real example output: [`docs/project-context.md`](docs/project-context.md)

---

## Features

- Git-style recursive directory tree (alphabetically sorted)
- File contents in syntax-highlighted code blocks
- Safe Markdown output (uses 4-backtick fences for `.md` files)
- Full `.gitignore` support including negation (`!`)
- Hard-coded ignores for common junk directories
- Per-file/folder content rules via `project-context.json`
- `--max-size`, `--truncate`, `--ext`, and `--verbose` flags
- Accurate token estimation
- `--stdout` mode for instant clipboard use
- Zero external dependencies (pure Go)

---

## Installation

### For end users

```bash
# Homebrew (recommended)
brew tap pixeljuggle/project-context
brew install project-context
```

### With Go

```bash
go install github.com/pixeljuggle/project-context/cmd/project-context@latest
```

### Download binaries

Pre-built binaries for Linux, macOS, and Windows (amd64 + arm64) are available on the [Releases page](https://github.com/pixeljuggle/project-context/releases).

---

## Usage

Run in the root of any project:

```bash
project-context
```

### Common workflows

```bash
# Most common — generate and copy straight to clipboard
project-context --stdout | pbcopy

# TypeScript / Next.js project (recommended defaults)
project-context --max-size 750 --truncate 150 --ext .ts,.tsx,.js,.jsx,.json,.md --stdout | pbcopy

# Skip tests and stories
project-context -I "*.test.*" -I "*.spec.*" -I "*.stories.*" --stdout | pbcopy
```

See the full generated example: [`docs/project-context.md`](docs/project-context.md)

---

## Configuration

Create `project-context.json` in your project root to set project-specific defaults:

```json
{
  "ignores": ["*.log", "*.tmp", "coverage/"],
  "rules": {
    "public/": { "content": false },
    "src/assets/": { "content": false }
  },
  "maxSizeKB": 500,
  "truncateLines": 150
}
```

Command-line flags always override config values.

---

## Command-Line Flags

| Flag             | Default                | Description                                        |
| ---------------- | ---------------------- | -------------------------------------------------- |
| `--stdout`       | false                  | Print to stdout instead of writing a file          |
| `--max-size`     | 1024                   | Maximum file size in KB (0 = unlimited)            |
| `--truncate`     | 0                      | Truncate large files to this many lines (0 = skip) |
| `--verbose`      | false                  | Print skipped/truncated files to stderr            |
| `--ext`          | (repeatable)           | Only include files with these extensions           |
| `--root`         | `.`                    | Project root directory                             |
| `--config`       | `project-context.json` | Path to config file                                |
| `--output`       | `project-context.md`   | Output filename                                    |
| `--no-gitignore` | false                  | Do not load `.gitignore`                           |
| `-I`             | (repeatable)           | Additional ignore pattern (supports `!` negation)  |
| `--version`      | —                      | Print version and exit                             |

---

## Development

### For Go developers

```bash
git clone https://github.com/pixeljuggle/project-context.git
cd project-context

make build          # builds to ./bin/project-context
make install        # installs globally
make run            # quick test run
```

### Makefile targets

| Command                | Purpose                         |
| ---------------------- | ------------------------------- |
| `make build`           | Build for current platform      |
| `make all`             | Build all supported platforms   |
| `make lint`            | Run gofmt, vet, and staticcheck |
| `make test`            | Run tests                       |
| `make release`         | Local snapshot release          |
| `make release-dry-run` | Validate release configuration  |
| `make clean`           | Remove build artifacts          |

---

## Contributing

Contributions are welcome. Please:

1. Fork and clone the repository.
2. Make your changes.
3. Run `make lint && make test`.
4. Submit a pull request.

---

## License

MIT

[Releases](https://github.com/pixeljuggle/project-context/releases) • [Issues](https://github.com/pixeljuggle/project-context/issues)
````

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

var Version = "dev" // overridden by ldflags

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var ignoreFlags, includeExts stringSlice

	rootDir := flag.String("root", ".", "Project root directory")
	configFile := flag.String("config", "project-context.json", "Path to JSON config")
	outputFile := flag.String("output", "project-context.md", "Output Markdown filename")
	noGitignore := flag.Bool("no-gitignore", false, "Skip loading .gitignore")
	stdout := flag.Bool("stdout", false, "Print to stdout (perfect for clipboard)")
	maxSizeKB := flag.Int("max-size", 1024, "Max file size in KB (0 = unlimited)")
	truncateLines := flag.Int("truncate", 0, "Truncate large files to N lines instead of skipping (0 = skip)")
	verbose := flag.Bool("verbose", false, "Verbose output (shows skipped/truncated files)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Var(&ignoreFlags, "I", "Additional ignore pattern (repeatable)")
	flag.Var(&includeExts, "ext", "Only include files with these extensions (repeatable, e.g. .go .ts)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generates a perfect project-context.md for LLMs.\n")
		fmt.Fprintf(os.Stderr, ".git is hard-coded ignored. Full .gitignore negation (!) supported.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample config (project-context.json):\n")
		fmt.Fprintf(os.Stderr, `{
  "ignores": ["*.log", "coverage/"],
  "rules": {"docs/": {"content": false}},
  "maxSizeKB": 500,
  "truncateLines": 200
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

	// Load config (now supports maxSizeKB + truncateLines)
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

	// Effective values (CLI flag wins; config provides sensible defaults)
	effectiveMaxSizeKB := *maxSizeKB
	if effectiveMaxSizeKB == 1024 && config.MaxSizeKB != 0 {
		effectiveMaxSizeKB = config.MaxSizeKB
	}
	effectiveTruncate := *truncateLines
	if effectiveTruncate == 0 && config.TruncateLines != 0 {
		effectiveTruncate = config.TruncateLines
	}

	maxSizeBytes := int64(0)
	if effectiveMaxSizeKB > 0 {
		maxSizeBytes = int64(effectiveMaxSizeKB) * 1024
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
	treeStr, contentFiles, err := generator.GenerateTreeAndFiles(root, ignorePatterns, config.Rules, includeExts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	mdContent := generator.BuildMarkdown(treeStr, contentFiles, root, maxSizeBytes, effectiveTruncate, *verbose)

	if *stdout {
		os.Stdout.WriteString(mdContent)
		fmt.Fprintf(os.Stderr, "Project context written to stdout (%d files)\n", len(contentFiles))
		return
	}

	outPath := *outputFile
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(root, outPath)
	}
	if err := os.WriteFile(outPath, []byte(mdContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created %s\n", outPath)
	fmt.Printf("   • Tree + %d file contents\n", len(contentFiles))
	fmt.Printf("   • .git always ignored • negation (!) supported\n")
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
	for p, rule := range rules {
		p = filepath.ToSlash(strings.TrimSuffix(p, "/"))
		if relPath == p || strings.HasPrefix(relPath, p+"/") {
			return rule.Content
		}
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
```

### internal/generator/generator_test.go

```go
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTreeAndFiles(t *testing.T) {
	tmp := t.TempDir()

	// Setup test project
	os.MkdirAll(filepath.Join(tmp, "src"), 0755)
	os.MkdirAll(filepath.Join(tmp, "docs"), 0755)
	os.WriteFile(filepath.Join(tmp, "src/main.go"), []byte("package main\nfunc main(){}"), 0644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Test\n"), 0644)
	os.WriteFile(filepath.Join(tmp, "docs/secret.md"), []byte("secret"), 0644)
	os.WriteFile(filepath.Join(tmp, ".git/config"), []byte("ignored"), 0644)
	os.WriteFile(filepath.Join(tmp, "node_modules/foo.js"), []byte("ignored"), 0644)

	tests := []struct {
		name             string
		ignores          []string
		rules            map[string]Rule
		includeExts      []string
		wantTreeContains []string // substrings that MUST appear in the tree
		wantTreeNot      []string // substrings that MUST NOT appear in the tree
		wantInContent    []string // files that should be in content list
		wantNotInContent []string // files that must NOT be in content list
	}{
		{
			name:             "basic tree + hard-coded ignores",
			ignores:          nil,
			rules:            nil,
			includeExts:      nil,
			wantTreeContains: []string{"README.md", "src/", "main.go", "docs/", "secret.md"},
			wantTreeNot:      []string{".git", "node_modules"},
			wantInContent:    []string{"src/main.go", "README.md", "docs/secret.md"},
			wantNotInContent: []string{".git/config", "node_modules/foo.js"},
		},
		{
			name:             "negation support (!)",
			ignores:          []string{"*.md", "!README.md"},
			rules:            nil,
			includeExts:      nil,
			wantTreeContains: []string{"README.md", "src/", "main.go"},
			wantTreeNot:      []string{".git", "node_modules", "secret.md", "docs/secret.md"},
			wantInContent:    []string{"src/main.go", "README.md"},
			wantNotInContent: []string{"docs/secret.md"},
		},
		{
			name:             "content rule + include ext",
			ignores:          nil,
			rules:            map[string]Rule{"docs/": {Content: false}},
			includeExts:      []string{".go"},
			wantTreeContains: []string{"README.md", "src/", "main.go", "docs/", "secret.md"},
			wantTreeNot:      []string{".git", "node_modules"},
			wantInContent:    []string{"src/main.go"},
			wantNotInContent: []string{"README.md", "docs/secret.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, files, err := GenerateTreeAndFiles(tmp, tt.ignores, tt.rules, tt.includeExts)
			if err != nil {
				t.Fatal(err)
			}

			// Verify tree output
			for _, want := range tt.wantTreeContains {
				if !strings.Contains(tree, want) {
					t.Errorf("expected %q in tree, but not found.\nTree was:\n%s", want, tree)
				}
			}
			for _, not := range tt.wantTreeNot {
				if strings.Contains(tree, not) {
					t.Errorf("%q should NOT be in tree", not)
				}
			}

			// Verify content files list
			for _, want := range tt.wantInContent {
				if !contains(files, want) {
					t.Errorf("missing %s in content files", want)
				}
			}
			for _, not := range tt.wantNotInContent {
				if contains(files, not) {
					t.Errorf("unexpected %s in content files", not)
				}
			}
		})
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
```

