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