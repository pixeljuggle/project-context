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
