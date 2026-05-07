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
