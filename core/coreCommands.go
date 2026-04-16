package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *EditorCore) cmdCommand(_ *EditorCore, arg []string) error {
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
	return nil
}

func (e *EditorCore) cmdSettings(*EditorCore, []string) error {
	e.loopChangeSettings()
	return nil
}

func (e *EditorCore) cmdSave(*EditorCore, []string) error {
	var saveBuffer []rune

	if e.SourceFile != "" {
		err := e.WriteBufferToFile(e.SourceFile)
		if err != nil {
			e.PrintMessageStyle(0, e.Rows, e.Styles.Error,
				fmt.Sprintf("Error saving file: %s", err.Error()))
			e.Terminal.Show()
			e.Terminal.PollEvent()
		}
		return nil
	}

	for {
		e.Terminal.Clear()
		e.DisplayBuffer()
		e.DisplayStatus()
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows / 2), e.Styles.Message, "Save As:")
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, e.Styles.Message, string(saveBuffer))
		e.Terminal.Show()

		event := e.Terminal.PollEvent()

		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEnter {
				filename := string(saveBuffer)
				if filename != "" {
					err := e.WriteBufferToFile(filename)
					if err != nil {
						e.PrintMessageStyle(0, e.Rows, e.Styles.Error,
							fmt.Sprintf("Error saving file: %s", err.Error()))
						e.Terminal.Show()
						e.Terminal.PollEvent()
					} else {
						return nil
					}
				}
				return nil
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(saveBuffer) > 0 {
					saveBuffer = saveBuffer[:len(saveBuffer)-1]
				}
			} else if ev.Key() == tcell.KeyEscape {
				return nil
			} else if ev.Rune() != 0 {
				saveBuffer = append(saveBuffer, ev.Rune())
			}
		}
	}
}

func (e *EditorCore) cmdSaveAs(*EditorCore, []string) error {
	var saveBuffer []rune
	bestGuess := ""

	updateGuess := func() {
		candidates, err := getFilesForPrefix(string(saveBuffer))
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
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows / 2), e.Styles.Message, "Save As:")

		guessStyle := e.Styles.Message.Attributes(tcell.AttrItalic)
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, guessStyle, string(bestGuess))
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, e.Styles.Message.Attributes(tcell.AttrBold), string(saveBuffer))
		e.Terminal.Show()

		event := e.Terminal.PollEvent()

		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEnter {
				filename := string(saveBuffer)
				if filename != "" {
					err := e.WriteBufferToFile(filename)
					if err != nil {
						e.PrintMessageStyle(0, e.Rows, e.Styles.Error,
							fmt.Sprintf("Error saving file: %s", err.Error()))
						e.Terminal.Show()
						e.Terminal.PollEvent()
					} else {
						return nil
					}
				}
				return nil
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(saveBuffer) > 0 {
					saveBuffer = saveBuffer[:len(saveBuffer)-1]
				}
				updateGuess()
			} else if ev.Key() == tcell.KeyEscape {
				return nil
			} else if ev.Key() == tcell.KeyTAB {
				if bestGuess != "" {
					saveBuffer = []rune(bestGuess)
					updateGuess()
				}
			} else if ev.Rune() != 0 {
				saveBuffer = append(saveBuffer, ev.Rune())
				updateGuess()
			}
		}
	}
}

func (e *EditorCore) cmdOpen(*EditorCore, []string) error {
	var openBuffer []rune
	bestGuess := ""

	updateGuess := func() {
		candidates, err := getFilesForPrefix(string(openBuffer))
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
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows / 2), e.Styles.Message, "Open File:")

		guessStyle := e.Styles.Message.Attributes(tcell.AttrItalic)
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, guessStyle, string(bestGuess))
		e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, e.Styles.Message.Attributes(tcell.AttrBold), string(openBuffer))
		e.Terminal.Show()

		event := e.Terminal.PollEvent()

		//Handle current input
		switch ev := event.(type) {
		case *tcell.EventKey:
			e.PrintMessageStyle((e.Cols/2)-e.LineCountWidth, (e.Rows/2)+1, e.Styles.Message, string(openBuffer))
			if ev.Key() == tcell.KeyEnter {
				filename := string(openBuffer)
				if filename != "" {
					newTEXTBUFFER, err := e.OpenFile(filename)
					if err != nil {
						e.PrintMessageStyle(0, e.Rows, e.Styles.Error, "Error opening file")
						e.Terminal.Show()
						e.Terminal.PollEvent()
						return nil
					}
					e.TextBuffer = newTEXTBUFFER
					e.SourceFile = filename
					e.Terminal.Clear()
					return nil
				}
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(openBuffer) > 0 {
					openBuffer = openBuffer[:len(openBuffer)-1]
				}
				updateGuess()
			} else if ev.Key() == tcell.KeyEscape {
				return nil
			} else if ev.Key() == tcell.KeyTAB {
				if bestGuess != "" {
					openBuffer = []rune(bestGuess)
					updateGuess()
				}
			} else if ev.Rune() != 0 {
				openBuffer = append(openBuffer, ev.Rune())
				updateGuess()
			}
		}
	}
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
