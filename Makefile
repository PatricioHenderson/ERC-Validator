.PHONY: help build build-helpers build-api build-validator build-web3 fmtvet tidy pre-commit docker-build docker-up docker-down docker-build-up test

help:
	@echo "Usage: make <target>"
	@echo "Available targets:"
	@echo "  help            Show this help message"
	@echo "  build           Compile all binaries (helpers, api, validator, web3, admin)"
	@echo "  build-helpers   Compile the helpers binary"
	@echo "  build-api       Compile the API binary"
	@echo "  build-validator Compile the validator binary"
	@echo "  build-web3      Compile the web3 binary"
	@echo "  build-admin     Compile the admin microservice binary"
	@echo "  fmtvet          Run 'go fmt' and 'go vet' on all services"
	@echo "  tidy            Run 'go mod tidy' on all services"
	@echo "  pre-commit      Runs tidy, fmtvet, and build (ideal for pre-commit hooks)"
	@echo "  docker-build    Run pre-commit and then build Docker images with no cache"
	@echo "  docker-up       Start all containers (docker-compose up)"
	@echo "  docker-down     Stop and remove all containers (docker-compose down)"
	@echo "  docker-build-up Build Docker images and start containers"
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

build-admin:
	@echo "Building admin..."
	cd services/admin/cmd && go build -o ../../../bin/admin

build: build-helpers build-api build-validator build-web3 build-admin

fmtvet:
	@echo "Formatting and vetting services/api..."
	cd services/api && go fmt ./... && go vet ./...
	@echo "Formatting and vetting services/helpers..."
	cd services/helpers && go fmt ./... && go vet ./...
	@echo "Formatting and vetting services/validator..."
	cd services/validator && go fmt ./... && go vet ./...
	@echo "Formatting and vetting services/web3..."
	cd services/web3 && go fmt ./... && go vet ./...
	@echo "Formatting and vetting services/admin..."
	cd services/admin && go fmt ./... && go vet ./...

tidy:
	@echo "Running go mod tidy in services/api..."
	cd services/api && go mod tidy
	@echo "Running go mod tidy in services/helpers..."
	cd services/helpers && go mod tidy
	@echo "Running go mod tidy in services/validator..."
	cd services/validator && go mod tidy
	@echo "Running go mod tidy in services/web3..."
	cd services/web3 && go mod tidy
	@echo "Running go mod tidy in services/admin..."
	cd services/admin && go mod tidy

pre-commit: tidy fmtvet build

docker-build: pre-commit
	docker-compose build --no-cache

docker-up:
	docker-compose up

docker-down:
	docker-compose down

docker-build-up: docker-build docker-up

test:
	@echo "Running tests in services/api..."
	cd services/api && go test ./... -v
	@echo "Running tests in services/helpers..."
	cd services/helpers && go test ./... -v
	@echo "Running tests in services/validator..."
	cd services/validator && go test ./... -v
	@echo "Running tests in services/web3..."
	cd services/web3 && go test ./... -v
	@echo "Running tests in services/admin..."
	cd services/admin && go test ./... -v