.PHONY: all clean build install

GOFLAGS ?= $(GOFLAGS:)

all: install

build:	clean
	@go build $(GOFLAGS) ./...

install:	build
	@go install $(GOFLAGS) ./...

clean:
	@go clean $(GOFLAGS) -i ./...
