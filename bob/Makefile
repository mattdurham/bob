# bob/Makefile — build the bob binary and WASM extensions.
#
# make             — build the bob binary → dist/bob
# make extensions  — build WASM extensions → dist/*.wasm
# make all         — build everything
# make clean       — remove dist/

DIST_DIR := dist
BINARY   := $(DIST_DIR)/bob

.PHONY: all build extensions clean

all: build

build: $(DIST_DIR)
	go build -o $(BINARY) ./cmd/
	@echo "Built $(BINARY)"

extensions: $(DIST_DIR)
	cd ../extensions/example && GOOS=wasip1 GOARCH=wasm go build -o $(CURDIR)/$(DIST_DIR)/hello.wasm .
	@echo "Built $(DIST_DIR)/hello.wasm"

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

clean:
	rm -rf $(DIST_DIR)
