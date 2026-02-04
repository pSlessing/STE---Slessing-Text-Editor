package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slessingTextEditor/core"
)

type GitPlugin struct {
	core *core.EditorCore
}

func (g *GitPlugin) Name() string {
	return "git"
}

func (g *GitPlugin) Commands() []core.Command {
	return []core.Command{
		{
			Name:        "gitpull",
			Aliases:     []string{"gpl"},
			Description: "Pull from repo",
			Execute:     g.PullCommand,
		},
		{
			Name:        "gitstatus",
			Aliases:     []string{"gs"},
			Description: "View Git Status",
			Execute:     g.StatusCommand,
		},
		{
			Name:        "gitpush",
			Aliases:     []string{"gps"},
			Description: "Pull from repo",
			Execute:     g.PushCommand,
		},
		{
			Name:        "gitcommit",
			Aliases:     []string{"gc"},
			Description: "Pull from repo",
			Execute:     g.CommitCommand,
		},
		{
			Name:        "gitbranch",
			Aliases:     []string{"gb"},
			Description: "Switch the branch you're currently on",
			Execute:     g.ChangeBranchCommand,
		},
	}
}

func (g *GitPlugin) Initialize(editorCore *core.EditorCore) error {
	g.core = editorCore
	return nil
}

func (g *GitPlugin) Cleanup() error {
	return nil
}

func (g *GitPlugin) PullCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = filepath.Dir(g.core.SourceFile)
	if g.core.SourceFile == "" {
		currentWd, _ := os.Getwd()
		cmd.Dir = filepath.Dir(currentWd)
	}
	out, err := cmd.CombinedOutput()

	message := string(out)
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error,
			"Git pull failed: "+message+" ("+err.Error()+")")
		g.core.Terminal.Show()
		return nil
	}

	g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Message, message)
	g.core.Terminal.Show()

	// Reload the file after pull
	openedTextBuffer, err := g.core.OpenFile(g.core.SourceFile)
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.TextBuffer = openedTextBuffer
	g.core.DisplayBuffer()
	g.core.DisplayStatus()
	g.core.Terminal.Show()
	return nil
}

func (g *GitPlugin) StatusCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "status")
	cmd.Dir = filepath.Dir(g.core.SourceFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(string(out))
	return nil
}

func (g *GitPlugin) CommitCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "commit", "-m", args[0])
	cmd.Dir = filepath.Dir(g.core.SourceFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(string(out))
	return nil
}

func (g *GitPlugin) PushCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "push")
	cmd.Dir = filepath.Dir(g.core.SourceFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage("result" + string(out))
	g.core.SetStatusMessage("Hi!")
	return nil
}

func (g *GitPlugin) ChangeBranchCommand(e *core.EditorCore, args []string) error {
	g.core.SetStatusMessage("Git branch command was ran")
	return nil
}

// CRITICAL: This exported function is how the plugin loader gets your plugin
func NewPlugin() core.Plugin {
	return &GitPlugin{}
}
