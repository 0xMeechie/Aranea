GO      := go
BIN_DIR := bin

.PHONY: all build cli runtime clean fmt tidy test install run-cli run-runtime

all: build

build: cli runtime

cli:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/aranea ./cmd/aranea

runtime:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/aranead ./cmd/aranead

run-cli: cli
	./$(BIN_DIR)/aranea

run-runtime: runtime
	./$(BIN_DIR)/aranead

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)

install: build
	cp $(BIN_DIR)/aranea /usr/local/bin/aranea
	cp $(BIN_DIR)/aranead /usr/local/bin/aranead
