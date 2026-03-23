VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.version=$(VERSION)
BINARY := jgc_exporter
DIST_DIR := dist
BUILD_GOOS := $(shell go env GOOS)
BUILD_GOARCH := $(shell go env GOARCH)
BUILD_SUFFIX := $(if $(filter windows,$(BUILD_GOOS)),.exe,)

.PHONY: build test clean cross dist-dir

build: dist-dir
	go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-$(BUILD_GOOS)-$(BUILD_GOARCH)$(BUILD_SUFFIX) .

all: dist-dir cross-linux cross-darwin cross-windows

dist-dir:
	@mkdir -p $(DIST_DIR)

test:
	go test ./... -v

clean:
	rm -rf $(DIST_DIR)
	rm -f $(BINARY) $(BINARY)-*

cross-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-linux-arm64 .

cross-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-darwin-arm64 .

cross-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-windows-amd64.exe .
