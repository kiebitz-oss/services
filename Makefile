.PHONY: all test clean build install examples secrets setup test-setup test-keys

SHELL := /bin/bash

GOFLAGS ?= -ldflags=\"-extldflags=-static\" $(GOFLAGS:)
BENCHARGS ?= -run=NONE -bench=. -benchmem

export KIEBITZ_TEST = yes

KIEBITZ_TEST_SETTINGS ?= "$(shell pwd)/settings/test"
KIEBITZ_ADMIN_SETTINGS ?= "$(KIEBITZ_TEST_SETTINGS)/002_admin.json"

all: dep install

build:
	CGO_ENABLED=0 go build $(GOFLAGS) ./...

dep:
	@go get ./...

install:
	CGO_ENABLED=0 go install $(GOFLAGS) ./...

test-setup: dep test-keys

test-keys:
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) kiebitz admin keys setup

test: test-setup
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) go test $(testargs) `go list ./...`

test-races: test-setup
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) go test -race $(testargs) `go list ./...`

bench: test-setup
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) go test $(BENCHARGS) $(GOFLAGS) `go list ./... | grep -v api/`

clean:
	@go clean $(GOFLAGS) -i ./...

copyright:
	python3 .scripts/make_copyright_headers.py

certs:
	rm -rf settings/dev/certs/*
	rm -rf settings/test/certs/*
	(cd settings/dev/certs; ../../../.scripts/make_certs.sh)
	(cd settings/test/certs; ../../../.scripts/make_certs.sh)
