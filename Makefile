# Project specific variables
PROJECT = reproxy
DESCRIPTION = Reverse Proxy

# --- the rest of the file should not need to be configured ---

# GO env
GOPATH=$(shell pwd)
GO=go
GOCMD=GOPATH=$(GOPATH) $(GO)

# Build versioning
VERSION = $(shell git describe --tags --always)
RELEASE = $(shell git log --pretty=oneline | wc -l | tr -d ' ')
COMMIT = $(shell git log -1 --format="%h" 2>/dev/null || echo "0")
BUILD_DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
FLAGS = -ldflags "\
  -X reproxy/constants.VERSION=$(VERSION) \
  -X reproxy/constants.COMMIT=$(COMMIT) \
  -X reproxy/constants.BUILD_DATE=$(BUILD_DATE) \
  "

GOBUILD = $(GOCMD) build $(FLAGS)

.PHONY: all clean build_one build_all setup run test coverage statics

all:	build_one

build_one: test
	$(GOBUILD) -o bin/$(PROJECT) $(PROJECT)

build_all: test
	@# https://golang.org/doc/install/source
	GOARCH=amd64 GOOS=linux   $(GOBUILD) -o bin/$(PROJECT).linux64 $(PROJECT)
	GOARCH=386   GOOS=linux   $(GOBUILD) -o bin/$(PROJECT).linux32 $(PROJECT)
	GOARCH=amd64 GOOS=darwin  $(GOBUILD) -o bin/$(PROJECT).mac64 $(PROJECT)
	GOARCH=386   GOOS=darwin  $(GOBUILD) -o bin/$(PROJECT).mac32 $(PROJECT)
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o bin/$(PROJECT).win64.exe $(PROJECT)
	GOARCH=386   GOOS=windows $(GOBUILD) -o bin/$(PROJECT).win32.exe $(PROJECT)

setup: clean
	$(GOCMD) get $(PROJECT)

clean:
	$(GOCMD) clean
	rm -fR $(GOBIN)
	rm -fR $(GOPATH)/pkg
	rm -fR .rpm
	rm -fR dist

run: build_one
	$(GOPATH)/bin/$(PROJECT)

test:
	$(GOCMD) test ./src/$(PROJECT)/... -cover

coverage:
	mkdir -p coverage
	$(GOCMD) test ./src/$(PROJECT)/db/metrics -cover -covermode=count -coverprofile=coverage/$(PROJECT).out
	$(GOCMD) tool cover -html=coverage/$(PROJECT).out

statics:
	$(GOCMD) run src/genstatic/genstatic.go --dir=static/ --package=files > src/reproxy/files/data.go
