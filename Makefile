.PHONY: all build build-plugins clean run

# Build everything
all: build-plugins build

# Build the main editor
build:
	go build -o ste main.go

# Build all plugins
build-plugins:
	@echo "Building plugins..."
	@mkdir -p modules
	@for plugin in modules/*Plugin*.go modules/*plugin*.go; do \
		if [ -f "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin .go); \
			echo "Building $$plugin_name.so..."; \
			go build -buildmode=plugin -o "modules/$$plugin_name.so" "$$plugin"; \
		fi \
	done
	@echo "Plugins built successfully!"

# Clean built files
clean:
	rm -f ste
	rm -f modules/*.so

# Run the editor
run: build-plugins build
	./ste