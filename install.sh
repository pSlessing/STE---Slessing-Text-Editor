#!/bin/bash

# STE Text Editor - Installation Script
# This script will build and install the STE text editor

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="ste"
INSTALL_DIR="/usr/local/bin"

echo -e "${GREEN}=== STE Text Editor Installation ===${NC}\n"

# Check if Go is installed | NEEDS REFACTORING; JUST RUN GO
#echo -e "${YELLOW}Checking for Go installation...${NC}"
#if ! command -v go &> /dev/null; then
#    echo -e "${RED}Error: Go is not installed. Please install Go 1.16 or higher.${NC}"
#    echo "Visit: https://golang.org/doc/install"
#    exit 1
#fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}Found Go: ${GO_VERSION}${NC}\n"

# Initialize Go module if go.mod doesn't exist
if [ ! -f "go.mod" ]; then
    echo -e "${YELLOW}Initializing Go module...${NC}"
    go mod init ste-text-editor
    echo -e "${GREEN}Go module initialized${NC}\n"
fi

# Download and organize dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod tidy
echo -e "${GREEN}Dependencies organized${NC}\n"

# Download all dependencies to local cache
echo -e "${YELLOW}Caching dependencies...${NC}"
go mod download
echo -e "${GREEN}Dependencies cached${NC}\n"

# Verify dependencies
echo -e "${YELLOW}Verifying dependencies...${NC}"
if go mod verify; then
    echo -e "${GREEN}Dependencies verified successfully${NC}\n"
else
    echo -e "${RED}Warning: Dependency verification failed${NC}\n"
fi

# Build the binary
echo -e "${YELLOW}Building ${BINARY_NAME}...${NC}"
go build -o "${BINARY_NAME}"
echo -e "${GREEN}Build successful${NC}\n"

# Install the binary
echo -e "${YELLOW}Installing ${BINARY_NAME} to ${INSTALL_DIR}...${NC}"
if [ -w "${INSTALL_DIR}" ]; then
    cp "${BINARY_NAME}" "${INSTALL_DIR}/"
else
    echo -e "${YELLOW}Requires elevated privileges...${NC}"
    sudo cp "${BINARY_NAME}" "${INSTALL_DIR}/"
fi
echo -e "${GREEN}Installed to ${INSTALL_DIR}/${BINARY_NAME}${NC}\n"

# Clean up local binary
echo -e "${YELLOW}Cleaning up...${NC}"
rm "${BINARY_NAME}"
echo -e "${GREEN}Local binary removed${NC}\n"

# Verify installation
if command -v "${BINARY_NAME}" &> /dev/null; then
    echo -e "${GREEN}=== Installation Complete ===${NC}"
    echo -e "You can now run '${BINARY_NAME}' from anywhere!"
    echo -e "\nTry it out: ${YELLOW}${BINARY_NAME}${NC}"
else
    echo -e "${RED}Warning: ${BINARY_NAME} was installed but is not in PATH${NC}"
    echo -e "You may need to add ${INSTALL_DIR} to your PATH"
fi