package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (e *EditorCore) cmdQuit(*EditorCore, []string) error {
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

func (e *EditorCore) cmdOpen(*EditorCore, []string) error {
	var openBuffer []rune

	bestGuess := ""

	//Get files in directory
	currentDirFiles, err := e.GetCurrentPathFiles()
	if err != nil {
		// Show error but continue with current buffer
		e.PrintMessageStyle(0, e.Rows, e.Styles.Error, "Error getting names of files")
		e.Terminal.Show()
		e.Terminal.PollEvent()
		return nil
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
						// Show error but continue with current buffer
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
				break
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(openBuffer) > 0 {
					openBuffer = openBuffer[:len(openBuffer)-1]
				}
			} else if ev.Key() == tcell.KeyEscape {
				return nil

			} else if ev.Key() == tcell.KeyTAB {
				openBuffer = []rune(bestGuess)
			} else if ev.Rune() != 0 {
				openBuffer = append(openBuffer, ev.Rune())
				//Find new best guess from current input
				for _, e := range currentDirFiles {
					if strings.HasPrefix(e, string(openBuffer)) {
						bestGuess = e
					}
				}
			}
		}
	}
}

func (*EditorCore) cmdHelp(*EditorCore, []string) error {
	//TODO: Insert some form of help here lmao
	return nil
}

func (e *EditorCore) cmdPlugins(*EditorCore, []string) error {
	commands := e.ListCommands()
	i := 0
	for i < len(commands) {
		e.PrintMessageStyle(e.Cols/2, i, e.Styles.Message, commands[i])
		i++
	}
	return nil
}
