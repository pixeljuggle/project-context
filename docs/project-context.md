# Project Context

**Estimated tokens:** ~10176

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
│   ├── project-context.md
│   └── roadmap.md
├── go.mod
├── internal/
│   └── generator/
│       ├── generator.go
│       └── generator_test.go
├── npm/
│   ├── README.md
│   ├── bin/
│   │   └── project-context.js
│   └── package.json
└── project-context.json
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

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          registry-url: "https://registry.npmjs.org"

      # 1. Build all binaries with GoReleaser
      - uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}

      # 2. Publish the 5 platform-specific optional packages
      - name: Publish platform packages to npm
        if: startsWith(github.ref, 'refs/tags/v')
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}

          publish_platform_pkg() {
            GOOS=$1
            GOARCH=$2
            NODE_OS=$3
            NODE_ARCH=$4

            PKG_NAME="@pixeljuggle/project-context-${GOOS}-${GOARCH}"
            DIR="dist/npm-${GOOS}-${GOARCH}"
            mkdir -p "$DIR"

            # GoReleaser now uses _v1 / _v8.0 suffixes → find the correct binary
            SRC_DIR="dist/project-context_${GOOS}_${GOARCH}"
            if [ -d "${SRC_DIR}_v8.0" ]; then
              SRC_DIR="${SRC_DIR}_v8.0"
            elif [ -d "${SRC_DIR}_v1" ]; then
              SRC_DIR="${SRC_DIR}_v1"
            fi

            # Copy the binary
            if [ "$GOOS" = "windows" ]; then
              cp "${SRC_DIR}/project-context.exe" "$DIR/" || true
            else
              cp "${SRC_DIR}/project-context" "$DIR/"
            fi

            # Create minimal package.json
            node -e "
              const fs = require('fs');
              const pkg = {
                name: '${PKG_NAME}',
                version: '${VERSION}',
                os: ['${NODE_OS}'],
                cpu: ['${NODE_ARCH}'],
                license: 'MIT',
                repository: {
                  type: 'git',
                  url: 'git+https://github.com/pixeljuggle/project-context.git'
                },
                files: ['project-context', 'project-context.exe'],
                publishConfig: { access: 'public' }
              };
              fs.writeFileSync('${DIR}/package.json', JSON.stringify(pkg, null, 2) + '\n');
              console.log('Created ${PKG_NAME}/package.json');
            "

            echo "Publishing ${PKG_NAME} v${VERSION}..."
            npm publish "./$DIR" --access public
          }

          # Publish all platforms
          publish_platform_pkg "darwin" "arm64" "darwin" "arm64"
          publish_platform_pkg "darwin" "amd64" "darwin" "x64"
          publish_platform_pkg "linux" "arm64" "linux" "arm64"
          publish_platform_pkg "linux" "amd64" "linux" "x64"
          publish_platform_pkg "windows" "amd64" "win32" "x64"

        # 3. Publish the main router package
      - name: Publish main npm package
        if: startsWith(github.ref, 'refs/tags/v')
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          cd npm

          # Update main package version + optionalDependencies
          npm --no-git-tag-version --allow-same-version version $VERSION
          node -e '
            const fs = require("fs");
            let pkg = JSON.parse(fs.readFileSync("package.json", "utf8"));
            const ver = "'$VERSION'";
            pkg.optionalDependencies = {
              "@pixeljuggle/project-context-darwin-arm64": ver,
              "@pixeljuggle/project-context-darwin-amd64": ver,
              "@pixeljuggle/project-context-linux-arm64": ver,
              "@pixeljuggle/project-context-linux-amd64": ver,
              "@pixeljuggle/project-context-windows-amd64": ver
            };
            fs.writeFileSync("package.json", JSON.stringify(pkg, null, 2) + "\n");
          '

          cp ../README.md ./
          cp ../LICENSE ./

          npm publish --access public
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
!npm/bin/
npm/bin/*
!npm/bin/project-context.js

# Test coverage
*.out
*.test
*.prof

# Dependency directories (if you ever vendor)
vendor/
node_modules/

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
  - id: project-context
    env:
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

brews:
  - repository:
      owner: pixeljuggle
      name: homebrew-project-context
    name: project-context
    homepage: https://github.com/pixeljuggle/project-context
    description: Zero-dependency CLI that generates a perfect project-context.md for LLMs, code reviews, or documentation.
    license: MIT
    install: |
      bin.install "project-context"
    test: |
      system "#{bin}/project-context --version"
    # Optional: explicitly put it in Formula/ directory
    directory: Formula
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
	go run ./cmd/project-context --config ./project-context.json

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

# === Bump version (recommended way to release) ===
# Usage: make bump-version VERSION=0.1.5
bump-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Usage: make bump-version VERSION=0.1.5"; \
		exit 1; \
	fi
	@echo "🔄 Bumping version to $(VERSION)..."
	@node -e "\
		const fs = require('fs'); \
		let pkg = JSON.parse(fs.readFileSync('npm/package.json', 'utf8')); \
		const ver = '$(VERSION)'; \
		pkg.version = ver; \
		pkg.optionalDependencies = { \
			'@pixeljuggle/project-context-darwin-arm64': ver, \
			'@pixeljuggle/project-context-darwin-amd64': ver, \
			'@pixeljuggle/project-context-linux-arm64': ver, \
			'@pixeljuggle/project-context-linux-amd64': ver, \
			'@pixeljuggle/project-context-windows-amd64': ver \
		}; \
		fs.writeFileSync('npm/package.json', JSON.stringify(pkg, null, 2) + '\n'); \
		console.log('✅ npm/package.json updated to ' + ver); \
	"
	project-context
	git add docs/project-context.md
	git add npm/package.json
	git commit -m "chore: bump version to v$(VERSION)"
	git tag "v$(VERSION)"
	@echo ""
	@echo "✅ Version bumped and tagged!"
	@echo "Now run:"
	@echo "   git push && git push --tags"

# Legacy alias (still works)
sync-npm-version: bump-version

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

> **Under heavy development** — Configuration, flags, and output format will likely change significantly in the coming weeks.  
> Feedback, issues, and suggestions are very welcome!

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

### npm / Bun (recommended for team projects & monorepos)

Add it as a dev dependency so every contributor automatically gets the CLI when they run `bun install` (or `npm install`):

```bash
bun add -d @pixeljuggle/project-context
# or
npm install --save-dev @pixeljuggle/project-context
```

The `project-context` command becomes available immediately in any `package.json` script and in `./node_modules/.bin/`.

### Homebrew (recommended for individual use on macOS/Linux)

```bash
# Homebrew (recommended)
brew tap pixeljuggle/homebrew-project-context
brew install project-context
```

**macOS note**: The first time you run `project-context` after installing via Homebrew, macOS Gatekeeper may show a security warning.  
To clear it, run this one-time command:

```bash
xattr -d com.apple.quarantine $(which project-context)
```

Or go to **System Settings → Privacy & Security** and click **Allow Anyway**.

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
  "truncateLines": 150,
  "output": "docs/project-context.md" // relative path; CLI --output always wins
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

| Command                 | Purpose                                                                      |
| ----------------------- | ---------------------------------------------------------------------------- |
| `make build`            | Build for current platform                                                   |
| `make all`              | Build all supported platforms                                                |
| `make lint`             | Run gofmt, vet, and staticcheck                                              |
| `make test`             | Run tests                                                                    |
| `make release`          | Local snapshot release                                                       |
| `make release-dry-run`  | Validate release configuration                                               |
| `make sync-npm-version` | Update `npm/package.json` version + optionalDependencies from latest git tag |
| `make clean`            | Remove build artifacts                                                       |

---

## Releasing

The release process is fully automated and uses a **single source of truth** (the git tag).

```bash
# Bump version, commit, tag, and push in one go
make bump-version VERSION=0.1.5
```

GitHub Actions + GoReleaser will automatically:

- Build native binaries for Linux, macOS, Windows (amd64 + arm64)
- Publish the 5 tiny platform-specific optional packages to npm
- Publish the main `@pixeljuggle/project-context` router package
- Create a GitHub Release with changelog
- Update the Homebrew formula in the tap

**Note:** `npm/package.json` now uses version `"0.0.0"` as a template. Never edit the version manually — always run `make sync-npm-version` (or let the CI handle it on tag push).

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
  "truncateLines": 200,
  "output": "project-context.md"
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
	// Effective output (CLI flag always wins)
	effectiveOutput := *outputFile
	if effectiveOutput == "project-context.md" && config.Output != "" {
		effectiveOutput = config.Output
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

	outPath := effectiveOutput
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

go 1.26
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
	"time"
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

// getProjectTitle returns the clean project name for metadata (same logic used for the tree root)
func getProjectTitle(root string) string {
	absRoot, _ := filepath.Abs(root)
	title := filepath.Base(absRoot)
	if title == "." || title == string(filepath.Separator) {
		title = "project"
	}
	return title
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

	// === YAML frontmatter metadata (always included) ===
	title := getProjectTitle(root)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05+00:00")

	md.WriteString(`---
type: 'Project Context'
title: ` + title + `
timestamp: ` + timestamp + `
---

`)

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

### npm/README.md

````md
# @pixeljuggle/project-context

**This is the npm distribution of [project-context](https://github.com/pixeljuggle/project-context).**

Install it as a dev dependency:

```bash
npm install --save-dev @pixeljuggle/project-context
# or
bun add -d @pixeljuggle/project-context
```

The `project-context` binary is automatically available in `./node_modules/.bin/`.

---

### Trusted Dependencies (npm v10+ / pnpm / Yarn)

npm now requires explicit trust for packages that ship native binaries (even with the modern optionalDependencies pattern we use).

Add this to your project's `package.json` to avoid security warnings:

```json
{
  "trustedDependencies": ["@pixeljuggle/project-context"]
}
```

This is a **one-time** setup and recommended for all users.

Full documentation → [GitHub README](https://github.com/pixeljuggle/project-context#readme)
````

### npm/bin/project-context.js

```js
#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const os = require("node:os");
const path = require("node:path");

const platform = os.platform();
const arch = os.arch();

// Map Node.js os/platform names → exact optional dependency package name
const knownPackages = {
  "darwin arm64": "@pixeljuggle/project-context-darwin-arm64",
  "darwin x64": "@pixeljuggle/project-context-darwin-amd64",
  "linux arm64": "@pixeljuggle/project-context-linux-arm64",
  "linux x64": "@pixeljuggle/project-context-linux-amd64",
  "win32 x64": "@pixeljuggle/project-context-windows-amd64",
};

const packageName = knownPackages[`${platform} ${arch}`];

if (!packageName) {
  console.error(`❌ Unsupported platform: ${platform} ${arch}`);
  process.exit(1);
}

try {
  // require.resolve gives us the exact location of the installed optional package
  const pkgPath = require.resolve(`${packageName}/package.json`);
  const binDir = path.dirname(pkgPath);

  const binaryName = platform === "win32" ? "project-context.exe" : "project-context";
  const binaryPath = path.join(binDir, binaryName);

  // Forward all arguments to the native binary
  const result = spawnSync(binaryPath, process.argv.slice(2), {
    stdio: "inherit",
    env: { ...process.env },
  });

  process.exit(result.status ?? 0);
} catch (error) {
  console.error(`❌ Failed to find or execute binary for ${packageName}.`);
  console.error(
    "This usually means the optional dependency was skipped (--no-optional) or failed to install.",
  );
  console.error("Try: npm install --include=optional");
  process.exit(1);
}
```

### npm/package.json

```json
{
  "name": "@pixeljuggle/project-context",
  "version": "0.1.8",
  "description": "Zero-dependency CLI that generates a perfect project-context.md for LLMs, code reviews, or documentation.",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/pixeljuggle/project-context.git"
  },
  "license": "MIT",
  "author": "alex",
  "bin": {
    "project-context": "bin/project-context.js"
  },
  "files": [
    "bin/project-context.js",
    "README.md",
    "LICENSE"
  ],
  "engines": {
    "node": ">=18"
  },
  "optionalDependencies": {
    "@pixeljuggle/project-context-darwin-arm64": "0.1.8",
    "@pixeljuggle/project-context-darwin-amd64": "0.1.8",
    "@pixeljuggle/project-context-linux-arm64": "0.1.8",
    "@pixeljuggle/project-context-linux-amd64": "0.1.8",
    "@pixeljuggle/project-context-windows-amd64": "0.1.8"
  }
}
```

### project-context.json

```json
{
  "rules": {
    "docs/project-context.md": { "content": false },
    "docs/roadmap.md": { "content": false },
    "npm/package-lock.json": { "content": false }
  },
  "output": "docs/project-context.md"
}
```

