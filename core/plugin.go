// core/plugin.go
package core

import (
	"fmt"
	"path/filepath"
	"plugin"
)

// Plugin interface that all plugins must implement
type Plugin interface {
	Name() string
	Commands() []Command
	Initialize(*EditorCore) error
	Cleanup() error
}

// Command represents a single command provided by a plugin
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Execute     func(*EditorCore, []string) error
}

// LoadPluginsFromDirectory automatically discovers and loads all .so plugins
func (e *EditorCore) LoadPluginsFromDirectory(dir string) error {
	// Find all .so files
	pattern := filepath.Join(dir, "*.so")
	plugins, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan plugin directory: %w", err)
	}

	fmt.Printf("Found %d plugins in %s\n", len(plugins), dir)

	// Load each plugin
	for _, pluginPath := range plugins {
		if err := e.loadPlugin(pluginPath); err != nil {
			// Don't fail completely if one plugin fails
			fmt.Printf("Warning: Failed to load plugin %s: %v\n", pluginPath, err)
			continue
		}
	}

	return nil
}

// loadPlugin loads a single plugin file
func (e *EditorCore) loadPlugin(path string) error {
	// Open the .so file
	plug, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for the required NewPlugin function
	symbolNewPlugin, err := plug.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin must export 'NewPlugin' function: %w", err)
	}

	// Assert the symbol is a function with the right signature
	newPluginFunc, ok := symbolNewPlugin.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin has incorrect signature")
	}

	// Create an instance of the plugin
	pluginInstance := newPluginFunc()

	// Register it with the editor
	return e.RegisterPlugin(pluginInstance)
}

// RegisterPlugin registers a plugin with the editor core
func (e *EditorCore) RegisterPlugin(p Plugin) error {
	// Initialize the plugin
	if err := p.Initialize(e); err != nil {
		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// Store the plugin
	e.plugins[p.Name()] = p

	// Register all commands from the plugin
	for _, cmd := range p.Commands() {
		// Register primary name
		e.commands[cmd.Name] = cmd

		// Register all aliases
		for _, alias := range cmd.Aliases {
			e.commands[alias] = cmd
		}

		fmt.Printf("Registered command: %s (plugin: %s)\n", cmd.Name, p.Name())
	}

	fmt.Printf("Successfully loaded plugin: %s\n", p.Name())
	return nil
}

// ExecuteCommand runs a command by name
func (e *EditorCore) ExecuteCommand(cmdName string, args []string) error {
	cmd, exists := e.commands[cmdName]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	return cmd.Execute(e, args)
}

// ListCommands returns all available commands
func (e *EditorCore) ListCommands() []string {
	commands := make([]string, 0, len(e.commands))
	seen := make(map[string]bool)

	for name, cmd := range e.commands {
		if !seen[cmd.Name] {
			commands = append(commands, name)
			seen[cmd.Name] = true
		}
	}

	return commands
}
