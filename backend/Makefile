# Developer tasks for calculator-api. Run "make help" for the list.

BINARY        := bin/calculator-api
COVER_PROFILE := coverage.out
COVER_HTML    := coverage.html
IMAGE         ?= calculator-api:dev
FUZZ_TIME     ?= 60s

.DEFAULT_GOAL := help
.PHONY: help run build test test-race cover fuzz lint fmt docker-build docker-run

help: ## Show the available tasks
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-13s %s\n", $$1, $$2}'

run: ## Run the service locally on :8080
	go run ./cmd/calculator-api

build: ## Build the binary into bin/
	go build -trimpath -o $(BINARY) ./cmd/calculator-api

test: ## Run all tests
	go test ./...

test-race: ## Run all tests with the race detector (needs cgo)
	go test -race ./...

cover: ## Write coverage.out and coverage.html and print the total
	go test -coverprofile=$(COVER_PROFILE) ./...
	go tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)
	go tool cover -func=$(COVER_PROFILE) | tail -n 1

fuzz: ## Fuzz the engine for FUZZ_TIME per target (default 60s)
	go test ./internal/engine -run '^$$' -fuzz '^FuzzEvaluate$$' -fuzztime $(FUZZ_TIME)
	go test ./internal/engine -run '^$$' -fuzz '^FuzzCalculate$$' -fuzztime $(FUZZ_TIME)

lint: ## Run golangci-lint with .golangci.yml
	golangci-lint run ./...

fmt: ## Format all Go code
	gofmt -w .

docker-build: ## Build the Docker image
	docker build -t $(IMAGE) .

docker-run: ## Run the Docker image on :8080
	docker run --rm -p 8080:8080 $(IMAGE)
