# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=portfolio-manager
BINARY_LINUX_X64=$(BINARY_NAME)_linux_amd64
BINARY_MAC_ARM=$(BINARY_NAME)_darwin_arm64
SOURCE_ENTRYPOINT=./cmd/portfolio

UI_PATH = web/ui
UI_NODE_MODULES_PATH = $(UI_PATH)/node_modules
REACT_APP_NPM_LICENSES_TARBALL = "npm_licenses.tar.bz2"

# Only build UI if PREBUILT_ASSETS_STATIC_DIR is not set
ifdef PREBUILT_ASSETS_STATIC_DIR
  SKIP_UI_BUILD = true
endif

# Set up conditional build tag for builtinassets.
# To build with embedded assets, run:
#   make build BUILTIN_ASSETS=1
ifneq ($(BUILTIN_ASSETS),)
    BUILD_TAGS=-tags "builtinassets"
else
    BUILD_TAGS=
endif

# All target
all: test build

# Build the project
build: swagger
	$(GOBUILD) $(BUILD_TAGS) -o $(BINARY_NAME) -v $(SOURCE_ENTRYPOINT)

# Build the project for multiple platforms
build-cross: swagger
	@echo '>> building assets for cross compilation'
	make build-linux-x64
	make build-mac-arm
	make build-windows-x64

.PHONY: ui-install
ui-install:
	cd $(UI_PATH) && npm install

.PHONY: ui-build
ui-build:
	cd $(UI_PATH) && CI="" npm run build

.PHONY: assets
ifndef SKIP_UI_BUILD
assets: ui-install ui-build

.PHONY: npm_licenses
npm_licenses: ui-install
	@echo ">> bundling npm licenses"
	rm -f $(REACT_APP_NPM_LICENSES_TARBALL) npm_licenses
	ln -s . npm_licenses
	find npm_licenses/$(UI_NODE_MODULES_PATH) -iname "license*" | tar cfj $(REACT_APP_NPM_LICENSES_TARBALL) --files-from=-
	rm -f npm_licenses
else
assets:
	@echo '>> skipping assets build, pre-built assets provided'

npm_licenses:
	@echo '>> skipping assets npm licenses, pre-built assets provided'
endif

.PHONY: assets-compress
assets-compress: assets
	@echo '>> compressing assets'
	scripts/compress_assets.sh

.PHONY: assets-tarball
assets-tarball: assets
	@echo '>> packaging assets'
	scripts/package_assets.sh

# Run tests
test: 
	$(GOTEST) ./...

test-verbose: 
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

swagger:
	swag init --quiet -g cmd/portfolio/main.go

# Cross compilation for Linux
build-linux-x64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_LINUX_X64) -v $(SOURCE_ENTRYPOINT)

# Cross compilation for macOS on ARM64:
build-mac-arm:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_MAC_ARM) -v $(SOURCE_ENTRYPOINT)

# Cross compilation for windows on AMD64
build-windows-x64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe -v $(SOURCE_ENTRYPOINT)

.PHONY: all build build-cross clean clean-db test run deps tidy build-linux-x64 build-mac-arm build-windows-x64 test test-verbose test-integration swagger
