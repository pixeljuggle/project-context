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
# 1. Sync version across npm/package.json (main + all optional deps)
make sync-npm-version

# 2. Validate everything
make release-dry-run

# 3. Tag and push (this triggers the full GitHub Actions release)
git tag v0.2.0          # bump to your next semantic version
git push && git push --tags
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
