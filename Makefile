.PHONY: build install clean test

BINARY_NAME=digital-history
INSTALL_PATH=/usr/local/bin

build:
	go build -o bin/$(BINARY_NAME) ./cmd/digital-history/

install: build
	cp bin/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	chmod +x $(INSTALL_PATH)/$(BINARY_NAME)

clean:
	rm -rf bin/
	go clean

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

run: build
	./bin/$(BINARY_NAME)

.DEFAULT_GOAL := build
