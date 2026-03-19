package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type EditorCore struct {
	// Buffer and cursor state
	TextBuffer [][]rune
	CursorX    int
	CursorY    int
	OffsetX    int
	OffsetY    int
	SourceFile string

	// Display state
	Terminal       tcell.Screen
	Styles         *StyleSet
	SettingsLength int
	Cols           int
	Rows           int

	// Plugin system
	plugins  map[string]Plugin
	commands map[string]Command

	// Other state
	InputBuffer    []rune
	LineCountWidth int

	//Constants, could maybe be moved into a settings/config file
	MaxWidth int

	statusMessage     string
	statusMessageTime time.Time
}

func NewEditor() (*EditorCore, error) {
	terminal, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := terminal.Init(); err != nil {
		return nil, err
	}

	editor := &EditorCore{
		TextBuffer:     [][]rune{{}},
		CursorX:        3,
		CursorY:        0,
		Terminal:       terminal,
		plugins:        make(map[string]Plugin),
		commands:       make(map[string]Command),
		LineCountWidth: 3,
		Styles: &StyleSet{
			Main:      tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault),
			Status:    tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorWhite),
			Message:   tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorWhite),
			Linecount: tcell.StyleDefault.Foreground(tcell.ColorDarkCyan).Background(tcell.ColorWhite),
			Error:     tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorRed),
		},
		SettingsLength:    10,
		MaxWidth:          78,
		statusMessage:     "",
		statusMessageTime: time.Now(),
	}
	settings, err := LoadSettings()
	if err != nil {
		editor.PrintMessageStyle(editor.Cols/2, editor.Rows/2, editor.Styles.Error, "Error loading settings"+(err.Error()))
	}

	editor.ApplySettings(settings)

	// Register built-in commands
	editor.registerBuiltInCommands()

	return editor, nil
}

func (e *EditorCore) registerBuiltInCommands() {
	builtins := []Command{
		{
			Name:        "quit",
			Aliases:     []string{"q"},
			Description: "Exit the editor",
			Execute:     e.cmdQuit,
		},
		{
			Name:        "write",
			Aliases:     []string{"w"},
			Description: "Enter write mode",
			Execute:     e.cmdWrite,
		},
		{
			Name:        "save",
			Aliases:     []string{"s"},
			Description: "Save current file",
			Execute:     e.cmdSave,
		},
		{
			Name:        "saveas",
			Aliases:     []string{"sa"},
			Description: "Save current file with a specific name",
			Execute:     e.cmdSaveAs,
		},
		{
			Name:        "open",
			Aliases:     []string{"o"},
			Description: "Open a file",
			Execute:     e.cmdOpen,
		},
		{
			Name:        "settings",
			Aliases:     []string{"se"},
			Description: "Change the settings",
			Execute:     e.cmdSettings,
		},
		{
			Name:        "help",
			Aliases:     []string{"h", "?"},
			Description: "Show available commands",
			Execute:     e.cmdHelp,
		},
		{
			Name:        "clear",
			Aliases:     []string{"c"},
			Description: "Clear the text buffer",
			Execute:     e.cmdClear,
		},
		{
			Name:        "plugins",
			Aliases:     []string{"p"},
			Description: "List the current plugins",
			Execute:     e.cmdPlugins,
		},
	}

	for _, cmd := range builtins {
		e.commands[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			e.commands[alias] = cmd
		}
	}
}

func (e *EditorCore) SetStatusMessage(msg string) {
	e.statusMessage = msg
	e.statusMessageTime = time.Now()
}

func (e *EditorCore) Run() {
	// Load plugins before starting
	if err := e.LoadPluginsFromDirectory("./modules"); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	e.mainLoop()
}

func (e *EditorCore) mainLoop() {
	e.CursorX = e.LineCountWidth
	for {
		e.Cols, e.Rows = e.Terminal.Size()
		//Ive forgotten why this is 2, one for buffer, but why another?
		//When 1, status bar is gone, so idk man
		e.Rows -= 2
		e.Cols -= e.LineCountWidth
		if e.Cols < e.MaxWidth {
			e.Cols = e.MaxWidth
		}
		e.Terminal.Clear()
		e.inputHandling()
		e.DisplayBuffer()
		e.DisplayStatus()
		e.Terminal.Show()
	}
}

func (e *EditorCore) handleCommand() {
	cmdText := string(e.InputBuffer)
	parts := strings.Fields(cmdText)

	if len(parts) == 0 {
		return
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	if err := e.ExecuteCommand(cmdName, args); err != nil {
		e.ShowError(err.Error())
	}
	e.Terminal.Clear()
	e.DisplayBuffer()
	e.DisplayStatus()
	e.Terminal.Show()
}

func (e *EditorCore) ShowError(err string) {
	e.PrintMessageStyle((e.Cols/2)-(len(err)/2), (e.Rows/2)-1, e.Styles.Error, "ERROR:")
	e.PrintMessageStyle((e.Cols/2)-(len(err)/2), e.Rows/2, e.Styles.Error, err)
}

func (e *EditorCore) inputHandling() {
	event := e.Terminal.PollEvent()

	switch ev := event.(type) {

	case *tcell.EventKey:
		mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
		if mod == tcell.ModNone {
			switch key {
			case tcell.KeyEnter:
				{
					e.handleCommand()
					e.InputBuffer = []rune{}
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				{
					if len(e.InputBuffer) > 0 {
						e.InputBuffer = e.InputBuffer[:len(e.InputBuffer)-1]
					}
				}
			case tcell.KeyEsc:
				{
					return
				}
			case tcell.KeyUp:
				{
					if e.CursorY > 0 {
						// Move cursor up within visible area
						e.CursorY--
					} else if e.OffsetY > 0 {
						// Scroll up when cursor is at top
						e.OffsetY--
					}
					// Adjust cursor X if moving to a shorter line
					if e.CursorY+e.OffsetY < len(e.TextBuffer) && e.CursorX-e.LineCountWidth > len(e.TextBuffer[e.CursorY+e.OffsetY]) {
						e.CursorX = len(e.TextBuffer[e.CursorY+e.OffsetY]) + e.LineCountWidth
					}
				}
			case tcell.KeyDown:
				{
					if e.CursorY < e.Rows-1 && e.CursorY+e.OffsetY+1 < len(e.TextBuffer) {
						// Move cursor down within visible area
						e.CursorY++
					} else if e.OffsetY+e.Rows < len(e.TextBuffer) {
						// Scroll down when cursor is at bottom
						e.OffsetY++
					}
					// Adjust cursor X if moving to a shorter line
					if e.CursorY+e.OffsetY < len(e.TextBuffer) && e.CursorX-e.LineCountWidth > len(e.TextBuffer[e.CursorY+e.OffsetY]) {
						e.CursorX = len(e.TextBuffer[e.CursorY+e.OffsetY]) + e.LineCountWidth
					}
				}
			case tcell.KeyLeft:
				{
					if e.CursorX > e.LineCountWidth {
						e.CursorX--
						// Horizontal scroll left if needed
						if e.CursorX < e.LineCountWidth {
							e.CursorX = e.LineCountWidth
						}
					} else if e.OffsetX > 0 {
						e.OffsetX--
					}
				}
			case tcell.KeyRight:
				{
					if e.CursorY+e.OffsetY < len(e.TextBuffer) {
						// Only allow moving right if not past end of line
						lineLen := len(e.TextBuffer[e.CursorY+e.OffsetY])
						if e.CursorX-e.LineCountWidth+e.OffsetX < lineLen {
							e.CursorX++
							// Horizontal scroll right if needed
							if e.CursorX >= e.Cols+e.LineCountWidth {
								e.OffsetX++
								e.CursorX = e.Cols + e.LineCountWidth - 1
							}
						}
					}
				}
			default:
				e.InputBuffer = append(e.InputBuffer, ch)
			}
		} else if mod == tcell.ModCtrl {
			switch key {
			case tcell.KeyLeft:
				{

				}
			case tcell.KeyRight:
				{

				}
			case tcell.KeyUp:
				{

				}
			case tcell.KeyDown:
				{

				}
			}
		} else if mod == tcell.ModAlt {
		}

	}
}
