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
