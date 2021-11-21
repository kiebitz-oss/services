.PHONY: all test clean build install examples secrets setup

SHELL := /bin/bash

GOFLAGS ?= $(GOFLAGS:)

export KIEBITZ_TEST = yes

KIEBITZ_TEST_SETTINGS ?= "$(shell pwd)/settings/test"

all: dep install

build:
	@go build $(GOFLAGS) ./...

dep:
	@go get ./...

setup: secrets

secrets:
	@printf "notification:\n\
  secret: `openssl rand -base64 32`\n\
appointments:\n \
  secret: `openssl rand -base64 32`\n\
	" > settings/dev/002_secrets.yml

install:
	@go install $(GOFLAGS) ./...

test: dep
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) go test $(testargs) `go list ./...`

test-races: dep
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) go test -race $(testargs) `go list ./...`

bench: dep
	KIEBITZ_SETTINGS=$(KIEBITZ_TEST_SETTINGS) go test -run=NONE -bench=. $(GOFLAGS) `go list ./... | grep -v api/`

clean:
	@go clean $(GOFLAGS) -i ./...

copyright:
	python3 .scripts/make_copyright_headers.py

certs:
	rm -rf settings/dev/certs/*
	rm -rf settings/test/certs/*
	(cd settings/dev/certs; ../../../.scripts/make_certs.sh)
	(cd settings/test/certs; ../../../.scripts/make_certs.sh)
