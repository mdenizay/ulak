BINARY  := ulak
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/mdenizay/ulak/cmd.Version=$(VERSION)"

.PHONY: build run test lint clean deps cross

build:
	go build $(LDFLAGS) -o $(BINARY) ./main.go

# Build a static binary for Linux deployment
cross:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
	go build $(LDFLAGS) -o $(BINARY)-linux-amd64 ./main.go

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
	rm -f $(BINARY) $(BINARY)-linux-amd64
