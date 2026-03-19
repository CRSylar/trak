.PHONY: build install clean

BINARY_DIR := $(HOME)/.local/bin
GOOS := darwin
GOARCH := arm64

build:
	@echo "Building trak and trakd for Apple Silicon..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/trak  ./cmd/trak
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/trakd ./cmd/trakd
	@echo "Done → bin/trak  bin/trakd"

install: build
	@echo "Installing to $(BINARY_DIR)..."
	@mkdir -p $(BINARY_DIR)
	cp bin/trak  $(BINARY_DIR)/trak
	cp bin/trakd $(BINARY_DIR)/trakd
	@echo "Installed. Make sure $(BINARY_DIR) is in your PATH."
	@echo ""
	@echo "Next: add the Raycast script:"
	@echo "  1. Open Raycast → Settings → Script Commands → Add Directory"
	@echo "  2. Point it to: $(PWD)/scripts/raycast"
	@echo "  3. Search 'Switch Work Project' in Raycast and assign a hotkey"

clean:
	rm -rf bin/

# Run locally for dev (native arch, no cross-compile)
dev:
	go build -o bin/trak  ./cmd/trak
	go build -o bin/trakd ./cmd/trakd
