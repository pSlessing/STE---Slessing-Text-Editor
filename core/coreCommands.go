package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *EditorCore) cmdCommand(_ *EditorCore, arg []string) error {
	if len(arg) == 0 {
		return fmt.Errorf("command requires an argument: cmd <command> [args...]")
	}
	cmd := exec.Command(arg[0], arg[1:]...)
	if cmd.Err != nil {
		e.PrintMessageStyle(e.Cols/2, e.Rows/2, e.Styles.Error, cmd.Err.Error())
		return nil
	}
	outputString, _ := cmd.Output()
	e.SetStatusMessage(string(outputString))
	return nil
}

func (e *EditorCore) cmdQuit(*EditorCore, []string) error {
	if e.Dirty {
		e.PrintMessageStyle((e.Cols/2)-20, e.Rows/2, e.Styles.Error,
			"Unsaved changes! Press q to quit anyway, any other key to cancel")
		e.Terminal.Show()
		event := e.Terminal.PollEvent()
		if ev, ok := event.(*tcell.EventKey); !ok || ev.Rune() != 'q' {
			return nil
		}
	}
	e.Terminal.Fini()
	os.Exit(0)
	return nil
}

func (e *EditorCore) cmdWrite(*EditorCore, []string) error {
	e.loopWrite()
	return nil
}

func (e *EditorCore) cmdClear(*EditorCore, []string) error {
	e.TextBuffer = [][]rune{{}}
	e.Dirty = true
	return nil
}

func (e *EditorCore) cmdSettings(*EditorCore, []string) error {
	e.loopChangeSettings()
	return nil
}

// promptForFilename draws a filename prompt (promptText) in the middle of
// the screen and reads input until Enter/Escape. If tabComplete is set, TAB
// accepts the best-guess completion from getFilesForPrefix. On Enter,
// onSubmit is called with the typed filename; if it returns an error, the
// error is shown and the prompt stays open for another attempt.
func (e *EditorCore) promptForFilename(promptText string, tabComplete bool, onSubmit func(filename string) error) error {
	var buffer []rune
	bestGuess := ""

	updateGuess := func() {
		if !tabComplete {
			return
		}
		candidates, err := getFilesForPrefix(string(buffer))
		if err != nil || len(candidates) == 0 {
			bestGuess = ""
			return
		}
		bestGuess = candidates[0]
	}

	for {
		e.Terminal.Clear()
		e.DisplayBuffer()
		e.DisplayStatus()
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows / 2), e.Styles.Message, promptText)

		if tabComplete {
			guessStyle := e.Styles.Message.Attributes(tcell.AttrItalic)
			e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, guessStyle, string(bestGuess))
			e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, e.Styles.Message.Attributes(tcell.AttrBold), string(buffer))
		} else {
			e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, e.Styles.Message, string(buffer))
		}
		e.Terminal.Show()

		event := e.Terminal.PollEvent()

		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEnter {
				filename := string(buffer)
				if filename == "" {
					return nil
				}
				if err := onSubmit(filename); err != nil {
					e.PrintMessageStyle(0, e.Rows, e.Styles.Error, err.Error())
					e.Terminal.Show()
					e.Terminal.PollEvent()
				} else {
					return nil
				}
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(buffer) > 0 {
					buffer = buffer[:len(buffer)-1]
				}
				updateGuess()
			} else if ev.Key() == tcell.KeyEscape {
				return nil
			} else if ev.Key() == tcell.KeyTAB && tabComplete {
				if bestGuess != "" {
					buffer = []rune(bestGuess)
					updateGuess()
				}
			} else if ev.Rune() != 0 {
				buffer = append(buffer, ev.Rune())
				updateGuess()
			}
		}
	}
}

func (e *EditorCore) cmdSave(*EditorCore, []string) error {
	if e.SourceFile != "" {
		if err := e.WriteBufferToFile(e.SourceFile); err != nil {
			e.PrintMessageStyle(0, e.Rows, e.Styles.Error,
				fmt.Sprintf("Error saving file: %s", err.Error()))
			e.Terminal.Show()
			e.Terminal.PollEvent()
		} else {
			e.Dirty = false
		}
		return nil
	}

	return e.cmdSaveAs(e, nil)
}

func (e *EditorCore) cmdSaveAs(*EditorCore, []string) error {
	return e.promptForFilename("Save As:", true, func(filename string) error {
		if err := e.WriteBufferToFile(filename); err != nil {
			return fmt.Errorf("error saving file: %s", err.Error())
		}
		e.Dirty = false
		return nil
	})
}

func (e *EditorCore) cmdOpen(*EditorCore, []string) error {
	return e.promptForFilename("Open File:", true, func(filename string) error {
		newTextBuffer, err := e.OpenFile(filename)
		if err != nil {
			return fmt.Errorf("error opening file: %s", err.Error())
		}
		e.TextBuffer = newTextBuffer
		e.SourceFile = filename
		e.Terminal.Clear()
		return nil
	})
}

func (e *EditorCore) cmdPlugins(*EditorCore, []string) error {
	stringToPrint := ""

	for _, v := range e.plugins {
		stringToPrint += (v.Name() + "\n")
	}
	stringToPrint = strings.TrimRight(stringToPrint, "\n")
	e.SetStatusMessage(stringToPrint)

	return nil
}

func (e *EditorCore) cmdHelp(*EditorCore, []string) error {
	commands := e.ListCommands()
	i := 0
	for i < len(commands) {
		e.PrintMessageStyle(e.Cols/2, i, e.Styles.Message, commands[i])
		i++
	}
	return nil
}
