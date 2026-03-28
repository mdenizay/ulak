BINARY  := ulak
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/mdenizay/ulak/cmd.Version=$(VERSION)"

.PHONY: build run test lint clean deps cross cross-all deb-amd64 deb-arm64 completions

build:
	go build $(LDFLAGS) -o $(BINARY) ./main.go

run:
	go run $(LDFLAGS) ./main.go $(ARGS)

test:
	go test ./... -count=1

test-verbose:
	go test ./... -v -count=1

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod download

clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64 $(BINARY)-linux-arm64
	rm -rf dist/ completions/

# ── Cross-compile (CGO_ENABLED=0 — modernc.org/sqlite, no gcc needed) ─────

cross-all: dist/ulak-linux-amd64 dist/ulak-linux-arm64

dist/ulak-linux-amd64:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build $(LDFLAGS) -o dist/ulak-linux-amd64 ./main.go

dist/ulak-linux-arm64:
	mkdir -p dist
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
	go build $(LDFLAGS) -o dist/ulak-linux-arm64 ./main.go

# ── Shell completions (requires local binary) ──────────────────────────────

completions: build
	mkdir -p completions
	./$(BINARY) completion bash > completions/ulak.bash
	./$(BINARY) completion zsh  > completions/ulak.zsh
	./$(BINARY) completion fish > completions/ulak.fish

# ── .deb packaging (requires nfpm: https://nfpm.goreleaser.com) ───────────

deb-amd64: dist/ulak-linux-amd64 completions
	GOARCH=amd64 VERSION=$(VERSION) nfpm package \
	  --config nfpm.yaml --packager deb \
	  --target dist/ulak_$(VERSION)_amd64.deb

deb-arm64: dist/ulak-linux-arm64 completions
	GOARCH=arm64 VERSION=$(VERSION) nfpm package \
	  --config nfpm.yaml --packager deb \
	  --target dist/ulak_$(VERSION)_arm64.deb
