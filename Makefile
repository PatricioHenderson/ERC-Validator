.PHONY: help build build-helpers build-api build-validator build-web3 lint tidy pre-commit docker-build docker-up docker-down docker-build-up test

help:
	@echo "Usage: make <target>"
	@echo "Available targets:"
	@echo "  help            Show this help message"
	@echo "  build           Compile all binaries (helpers, api, validator, web3)"
	@echo "  build-helpers   Compile the helpers binary"
	@echo "  build-api       Compile the API binary"
	@echo "  build-validator Compile the validator binary"
	@echo "  build-web3      Compile the web3 binary"
	@echo "  lint            Run golangci-lint with --fix on all services"
	@echo "  tidy            Run 'go mod tidy' on all services"
	@echo "  pre-commit      Runs tidy, lint, and build (ideal for pre-commit hooks)"
	@echo "  docker-build    Run pre-commit (tidy, lint, build) and then build Docker images with no cache"
	@echo "  docker-up       Start all containers (docker-compose up)"
	@echo "  docker-down     Stop and remove all containers (docker-compose down)"
	@echo "  docker-build-up Build Docker images and start containers (runs docker-build then docker-up)"
	@echo "  test            Run unit tests on all services"

build-helpers:
	@echo "Building helpers..."
	cd services/helpers/cmd && go build -o ../../../bin/helpers

build-api:
	@echo "Building API..."
	cd services/api/cmd && go build -o ../../../bin/api

build-validator:
	@echo "Building validator..."
	cd services/validator/cmd && go build -o ../../bin/validator

build-web3:
	@echo "Building web3..."
	cd services/web3/cmd && go build -o ../../bin/web3

build: build-helpers build-api build-validator build-web3

lint:
	@echo "Linting services/api..."
	cd services/api && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix ./...
	@echo "Linting services/helpers..."
	cd services/helpers && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix ./...
	@echo "Linting services/validator..."
	cd services/validator && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix ./...
	@echo "Linting services/web3..."
	cd services/web3 && go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix ./...

tidy:
	@echo "Running go mod tidy in services/helpers..."
	cd services/helpers && go mod tidy
	@echo "Running go mod tidy in services/api..."
	cd services/api && go mod tidy
	@echo "Running go mod tidy in services/validator..."
	cd services/validator && go mod tidy
	@echo "Running go mod tidy in services/web3..."
	cd services/web3 && go mod tidy

pre-commit: tidy lint build

docker-build: pre-commit
	docker-compose build --no-cache

docker-up:
	docker-compose up

docker-build-up: docker-build docker-up

docker-down:
	docker-compose down

test:
	@echo "Running unit tests..."
	cd services/api && go test ./...
	cd services/helpers && go test ./...
	cd services/validator && go test ./...
	cd services/web3 && go test ./...
