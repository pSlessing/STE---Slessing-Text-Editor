package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slessingTextEditor/core"
	"strings"
)

type GitPlugin struct {
	core *core.EditorCore
	wd   string
}

func (g *GitPlugin) Name() string {
	return "git"
}

func (g *GitPlugin) Commands() []core.Command {
	return []core.Command{
		{
			Name:        "gitpull",
			Aliases:     []string{"gpl"},
			Description: "Pull from remote",
			Execute:     g.PullCommand,
		},
		{
			Name:        "gitstatus",
			Aliases:     []string{"gs"},
			Description: "View git status",
			Execute:     g.StatusCommand,
		},
		{
			Name:        "gitpush",
			Aliases:     []string{"gps"},
			Description: "Push to remote",
			Execute:     g.PushCommand,
		},
		{
			Name:        "gitcommit",
			Aliases:     []string{"gc"},
			Description: "Commit with a message: gc <message>",
			Execute:     g.CommitCommand,
		},
		{
			Name:        "gitbranch",
			Aliases:     []string{"gb"},
			Description: "List branches (no args), switch branch (gb <name>), or create (gb -b <name>)",
			Execute:     g.ChangeBranchCommand,
		},
		{
			Name:        "gitadd",
			Aliases:     []string{"ga"},
			Description: "Stage files: ga <file> or ga for all",
			Execute:     g.AddCommand,
		},
		{
			Name:        "gitlog",
			Aliases:     []string{"gl"},
			Description: "Show recent commits",
			Execute:     g.LogCommand,
		},
		{
			Name:        "gitstash",
			Aliases:     []string{"gst"},
			Description: "Stash current changes",
			Execute:     g.StashCommand,
		},
		{
			Name:        "gitstashpop",
			Aliases:     []string{"gstp"},
			Description: "Pop the latest stash",
			Execute:     g.StashPopCommand,
		},
	}
}

func (g *GitPlugin) Initialize(editorCore *core.EditorCore) error {
	g.core = editorCore
	if g.core.SourceFile == "" {
		wdString, _ := os.Getwd()
		g.wd = wdString + "/"
	} else {
		g.wd = g.core.SourceFile
	}

	// Populate branch name in status bar
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.Output()
	if err == nil {
		g.core.CurrentBranch = strings.TrimSpace(string(out))
	}

	return nil
}

func (g *GitPlugin) Cleanup() error {
	return nil
}

func (g *GitPlugin) PullCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()

	message := strings.TrimSpace(string(out))
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error,
			"Git pull failed: "+message+" ("+err.Error()+")")
		g.core.Terminal.Show()
		return nil
	}

	g.core.SetStatusMessage(message)

	if g.core.SourceFile != "" {
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
	}
	return nil
}

func (g *GitPlugin) StatusCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "status")
	cmd.Dir = filepath.Dir(g.wd)
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
	if len(args) == 0 {
		g.core.SetStatusMessage("gitcommit requires a message: gc <message>")
		return nil
	}

	commitMSG := strings.Join(args, " ")
	cmd := exec.Command("git", "commit", "-a", "-m", commitMSG)
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(strings.TrimSpace(string(out)))
	return nil
}

func (g *GitPlugin) PushCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "push")
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(strings.TrimSpace(string(out)))
	return nil
}

func (g *GitPlugin) ChangeBranchCommand(e *core.EditorCore, args []string) error {
	// No args: list branches as overlay
	if len(args) == 0 {
		cmd := exec.Command("git", "branch")
		cmd.Dir = filepath.Dir(g.wd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
			g.core.Terminal.Show()
			return nil
		}
		g.core.SetStatusMessage(string(out))
		return nil
	}

	// -b <name>: create and switch to new branch
	if args[0] == "-b" {
		if len(args) < 2 {
			g.core.SetStatusMessage("Usage: gb -b <branch-name>")
			return nil
		}
		cmd := exec.Command("git", "checkout", "-b", args[1])
		cmd.Dir = filepath.Dir(g.wd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			g.core.SetStatusMessage(string(out))
			g.core.Terminal.Show()
			return nil
		}
		g.core.CurrentBranch = args[1]
		g.core.SetStatusMessage(strings.TrimSpace(string(out)))
		return nil
	}

	// <name>: switch branch
	cmd := exec.Command("git", "checkout", args[0])
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.CurrentBranch = args[0]
	g.core.SetStatusMessage(strings.TrimSpace(string(out)))
	return nil
}

func (g *GitPlugin) AddCommand(e *core.EditorCore, args []string) error {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}
	cmd := exec.Command("git", "add", target)
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	msg := strings.TrimSpace(string(out))
	if msg == "" {
		msg = "Staged: " + target
	}
	g.core.SetStatusMessage(msg)
	return nil
}

func (g *GitPlugin) LogCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "log", "--oneline", "-20")
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(string(out))
	g.core.Terminal.Show()
	return nil
}

func (g *GitPlugin) StashCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "stash")
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(strings.TrimSpace(string(out)))
	return nil
}

func (g *GitPlugin) StashPopCommand(e *core.EditorCore, args []string) error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = filepath.Dir(g.wd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g.core.PrintMessageStyle(g.core.Cols/2, g.core.Rows/2, g.core.Styles.Error, err.Error())
		g.core.Terminal.Show()
		return nil
	}
	g.core.SetStatusMessage(strings.TrimSpace(string(out)))
	return nil
}

// CRITICAL: This exported function is how the plugin loader gets your plugin
func NewPlugin() core.Plugin {
	return &GitPlugin{}
}
