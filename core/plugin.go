// core/plugin.go
package core

import (
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
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

// LoadPluginsFromDirectory automatically discovers and loads all .so plugins.
// The terminal is already in raw/alt-screen mode by the time this runs, so
// diagnostics are surfaced via the status message rather than printed to
// stdout (which would corrupt the display instead of being seen).
func (e *EditorCore) LoadPluginsFromDirectory(dir string) error {
	// Find all .so files
	pattern := filepath.Join(dir, "*.so")
	plugins, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan plugin directory: %w", err)
	}

	loaded := 0
	var warnings []string

	// Load each plugin
	for _, pluginPath := range plugins {
		if err := e.loadPlugin(pluginPath); err != nil {
			// Don't fail completely if one plugin fails
			warnings = append(warnings, fmt.Sprintf("Failed to load plugin %s: %v", filepath.Base(pluginPath), err))
			continue
		}
		loaded++
	}

	if len(plugins) > 0 {
		messages := append([]string{fmt.Sprintf("Loaded %d/%d plugins", loaded, len(plugins))}, warnings...)
		e.SetStatusMessage(strings.Join(messages, "\n"))
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
	}

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
