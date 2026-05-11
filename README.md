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

## Releasing

To create a new release:

```bash
# 1. Validate everything
make release-dry-run

# 2. Tag and push (this triggers the full automated release)
git tag v0.0.X
git push origin v0.0.X
```

GitHub Actions + GoReleaser will automatically:

- Build binaries for Linux, macOS, Windows (amd64 + arm64)
- Create a GitHub Release with changelog
- Update the Homebrew formula in the tap

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
