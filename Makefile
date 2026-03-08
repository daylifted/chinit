BINARY    := sh
GOOS      ?= linux
GOARCH    ?= amd64
LDFLAGS   := -s -w

.PHONY: all build test clean

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test:
	go test -v -race ./...

clean:
	rm -f $(BINARY)
