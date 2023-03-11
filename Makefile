.DEFAULT_GOAL := all

ROOT_PATH := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
TESTS_PATH := $(ROOT_PATH).tests/

clean:
	@rm -rf $(TESTS_PATH)

dirs:
	@mkdir -p $(TESTS_PATH)

test: dirs
	@go test -v ./... -coverprofile $(TESTS_PATH)cp.out
	@go tool cover -func=$(TESTS_PATH)cp.out -o $(TESTS_PATH)coverage.txt
	@go tool cover -html=$(TESTS_PATH)cp.out -o $(TESTS_PATH)coverage.html

lint:
	@golangci-lint run

.PHONY: all
all: lint test
