.DEFAULT_GOAL := all

ROOT_PATH := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
COVERAGE_PATH := $(ROOT_PATH).coverage

lint:
	@golangci-lint run

test:
	@rm -rf $(COVERAGE_PATH)
	@mkdir -p $(COVERAGE_PATH)
	@go test -v -coverpkg=./... ./... -coverprofile $(COVERAGE_PATH)/cp.out
	@go tool cover -func=$(COVERAGE_PATH)/cp.out -o $(COVERAGE_PATH)/coverage.txt
	@go tool cover -html=$(COVERAGE_PATH)/cp.out -o $(COVERAGE_PATH)/coverage.html

all: lint test

.PHONY: lint test all
