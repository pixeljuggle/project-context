# Project Context CLI

**A fast, zero-dependency CLI that generates perfect `project-context.md` files for LLMs.**

Feed your entire codebase to Claude, Cursor, Grok, Aider, GPT-4o, or any other AI coding agent with one command.

---

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

---

## Installation

### Option 1: Install globally (recommended)

```bash
go install github.com/pixeljuggle/project-context/cmd/project-context@latest
```

### Option 2: Build from source

```bash
git clone https://github.com/pixeljuggle/project-context.git
cd project-context
make install
```

Now you can run `project-context` from anywhere.

---

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

---

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

---

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
