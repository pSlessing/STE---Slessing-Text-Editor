# STE Text Editor

A basic text editor for the terminal, under active development

## Features

### Current Features
- **Command-based interface** - Main loop with commands executed via status bar
- **File operations** - Open, Save, and SaveAs commands for file management
- **Customizability** - Customizable color schemes with session persistence - Fuzzy file search within current directory
- **Autocomplete** - Autocomplete when opening files using tab to complete to the current guess

### Upcoming Features
- Syntax highlighting for multiple programming languages
- Command aliasing for improved workflow efficiency
- Persistent cursor position across mode transitions

## Installation

### Prerequisites
- Go 1.16 or higher

### Building from Source

0. Psst, you can use the install script ```make```

1. Clone the repository and navigate to the project directory

2. Initialize dependencies:
```bash
go mod tidy
```

3. Build the binary:
```bash
go build -o ste
```

4. Install system-wide (requires sudo):
```bash
sudo cp ste /usr/local/bin/
```

5. Clean up the local binary (optional):
```bash
rm ste
```

You can now run `ste` from anywhere in your terminal!

## Development

### Running Without Installing

Run directly using Go:
```bash
go run .
```

### Dependency Management

Download all dependencies:
```bash
go mod download
```

Verify dependency integrity:
```bash
go mod verify
```

small change for test
