BINARY := pgvet
BUILD_DIR := ./bin
CMD := ./cmd/pgvet

dependencies:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.11.2

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) $(CMD)

.PHONY: test
test:
	go test ./...

lint:
	go vet ./...
	PATH=${PWD}/bin:${PATH} golangci-lint -c ./configs/.golangci.yaml run

clean:
	rm -rf $(BUILD_DIR)

run:
	go run $(CMD) $(ARGS)
