GO      := go
BIN_DIR := bin

.PHONY: all build cli gateway runtime clean fmt tidy test install

all: build

build: cli gateway runtime

cli:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/aranea ./src/cmd/araneacli

gateway:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/aranea-gateway ./src/cmd/gateway

runtime:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/aranead ./src/cmd/aranead

run-cli: cli
	./$(BIN_DIR)/aranea

run-gateway: gateway
	./$(BIN_DIR)/aranea-gateway

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
	cp $(BIN_DIR)/aranea-gateway /usr/local/bin/aranea-gateway
