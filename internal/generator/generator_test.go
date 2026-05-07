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
	os.WriteFile(filepath.Join(tmp, ".git/config"), []byte("ignored"), 0644) // should be ignored
	os.WriteFile(filepath.Join(tmp, "node_modules/foo.js"), []byte("ignored"), 0644)

	tests := []struct {
		name         string
		ignores      []string
		rules        map[string]Rule
		includeExts  []string
		wantContains []string
		wantNot      []string
	}{
		{
			name:         "basic tree + hard-coded ignores",
			ignores:      nil,
			rules:        nil,
			includeExts:  nil,
			wantContains: []string{"src/main.go", "README.md"},
			wantNot:      []string{".git", "node_modules"},
		},
		{
			name:         "negation support (!)",
			ignores:      []string{"*.md", "!README.md"},
			rules:        nil,
			wantContains: []string{"README.md"},
			wantNot:      []string{"docs/secret.md"},
		},
		{
			name:         "content rule + include ext",
			ignores:      nil,
			rules:        map[string]Rule{"docs/": {Content: false}},
			includeExts:  []string{".go"},
			wantContains: []string{"src/main.go"},
			wantNot:      []string{"README.md", "docs/secret.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, files, err := GenerateTreeAndFiles(tmp, tt.ignores, tt.rules, tt.includeExts)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(tree, want) && !contains(files, want) {
					t.Errorf("missing %s in tree/files", want)
				}
			}
			for _, not := range tt.wantNot {
				if strings.Contains(tree, not) || contains(files, not) {
					t.Errorf("unexpected %s found", not)
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
