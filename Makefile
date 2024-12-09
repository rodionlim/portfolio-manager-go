# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=portfolio-manager
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_MAC_ARM=$(BINARY_NAME)_mac_arm64

# All target
all: test build

# Build the project
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/portfolio

# Run tests
test: 
	$(GOTEST) -v ./...

test-integration:
	$(GOTEST) -v -tags=integration -run Integration ./...

# Clean build files
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_MAC_ARM)
	rm -f *.log

# Wipe the database
clean-db:
	rm -rf *.db

# Run the application
run: build
	@./$(BINARY_NAME)

# Install dependencies
deps: 
	$(GOGET) -u ./...

# Tidy up the go.mod and go.sum files
tidy:
	$(GOCMD) mod tidy

# Cross compilation for Linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

# Cross compilation for macOS on ARM64build-mac-arm:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_MAC_ARM) -v

.PHONY: all build clean test run deps tidy build-linux build-mac-arm
