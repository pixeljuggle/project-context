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
	var ignoreFlags stringSlice

	rootDir := flag.String("root", ".", "Project root directory")
	configFile := flag.String("config", "project-context.json", "Path to JSON config (optional)")
	outputFile := flag.String("output", "project-context.md", "Output Markdown filename")
	noGitignore := flag.Bool("no-gitignore", false, "Skip loading .gitignore")
	stdout := flag.Bool("stdout", false, "Print to stdout instead of writing file")
	maxSizeKB := flag.Int("max-size", 1024, "Max file size in KB (0 = unlimited)")
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

	// Generate tree + content file list
	treeStr, contentFiles, err := generator.GenerateTreeAndFiles(root, ignorePatterns, config.Rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	maxSizeBytes := int64(0)
	if *maxSizeKB > 0 {
		maxSizeBytes = int64(*maxSizeKB) * 1024
	}

	mdContent := generator.BuildMarkdown(treeStr, contentFiles, root, maxSizeBytes)

	if *stdout {
		os.Stdout.WriteString(mdContent)
		fmt.Fprintf(os.Stderr, "✅ Project context written to stdout (%d files processed)\n", len(contentFiles))
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

	fmt.Printf("✅ Successfully created %s\n", outPath)
	fmt.Printf("   • Tree + %d file contents included\n", len(contentFiles))
	fmt.Printf("   • .git is always ignored (hard-coded)\n")
}
