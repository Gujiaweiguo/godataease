.PHONY: build test lint run clean fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary
BINARY_NAME=dataease-backend
BINARY_UNIX=$(BINARY_NAME)_unix

# Main package
MAIN_PACKAGE=./cmd/api

# Build
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Test
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Lint
lint:
	golangci-lint run ./...

# Vet
vet:
	$(GOCMD) vet ./...

# Format
fmt:
	$(GOCMD) fmt ./...

# Run
run:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PACKAGE) && ./$(BINARY_NAME)

# Clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Tidy
tidy:
	$(GOMOD) tidy

# All checks
check: fmt vet lint test

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME) $(MAIN_PACKAGE)
