.PHONY: build clean test fmt vet lint install vendor

BINARY_NAME=protonvpn-wg-config-generate
BUILD_DIR=build
CMD_DIR=cmd/protonvpn-wg

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "  Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@echo "  Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)/main.go
	@echo "  Linux arm..."
	@GOOS=linux GOARCH=arm go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm $(CMD_DIR)/main.go
	@echo "  macOS amd64..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	@echo "  macOS arm64..."
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/main.go
	@echo "  Windows amd64..."
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go
	@echo "  Windows arm64..."
	@GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(CMD_DIR)/main.go
	@echo "Done!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Update vendor directory
vendor:
	@echo "Updating vendor..."
	@go mod vendor

# Run the application
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Development build with race detector
dev:
	@echo "Building with race detector..."
	@go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(CMD_DIR)/main.go