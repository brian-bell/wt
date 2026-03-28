BIN_DIR  = bin
BINARY   = $(BIN_DIR)/wtui

.PHONY: build test run clean tidy

build:
	go build -o $(BINARY) ./cmd/wtui

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -rf $(BIN_DIR)

tidy:
	go mod tidy
