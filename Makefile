.PHONY: build install clean

BINARY_NAME=worklog
BUILD_DIR=.
INSTALL_DIR=$(HOME)/.local/bin

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@ln -sf $(CURDIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Symlink created: $(INSTALL_DIR)/$(BINARY_NAME) -> $(CURDIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

clean:
	@echo "Cleaning..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

help:
	@echo "Available targets:"
	@echo "  build   - Build the noter binary"
	@echo "  install - Build and create symlink in ~/.local/bin"
	@echo "  clean   - Remove binary and symlink"
	@echo "  help    - Show this help message"
