package core

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// PrintMessage #TODO: should this be able to use any, or standard colors every time?
func (e *EditorCore) PrintMessage(col, row int, fg, bg tcell.Color, msg string) {
	for _, c := range msg {
		currStyle := tcell.StyleDefault.Foreground(fg).Background(bg)
		e.Terminal.SetContent(col, row, c, nil, currStyle)
		col += runewidth.RuneWidth(c)
	}
}

func (e *EditorCore) PrintMessageStyle(col, row int, style tcell.Style, msg string) {
	for _, c := range msg {
		e.Terminal.SetContent(col, row, c, nil, style)
		col += runewidth.RuneWidth(c)
	}
}

func (e *EditorCore) DisplayBuffer() {
	var row, col int

	for row = 0; row <= e.Rows; row++ {
		textBufferRow := row + e.OffsetY

		e.DisplayLineNumber(row, textBufferRow)

		for col = 0; col < e.Cols; col++ {
			textBufferCol := col + e.OffsetX

			if textBufferRow >= 0 &&
				textBufferRow < len(e.TextBuffer) &&
				textBufferCol < len(e.TextBuffer[textBufferRow]) {
				e.Terminal.SetContent(col+e.LineCountWidth, row,
					e.TextBuffer[textBufferRow][textBufferCol],
					nil, e.Styles.Main)
			}
		}
	}
}

// TODO: IMPLEMENT MESSAGES IN STATUS BAR OR SOMETHING ELSE

/* type EditorCore struct {
    // ... existing fields
    statusMessage   string
    statusMessageTime time.Time
}

func (e *EditorCore) SetStatusMessage(msg string) {
    e.statusMessage = msg
    e.statusMessageTime = time.Now()
}

func (e *EditorCore) DisplayStatus() {
    // ... existing status display code

    // Show message if it's recent (e.g., within 3 seconds)
    if time.Since(e.statusMessageTime) < 3*time.Second {
        // Display e.statusMessage in status bar or separate line
    } else {
        e.statusMessage = "" // Clear old message
    }
} */

func (e *EditorCore) DisplayStatus() {
	var col int

	e.Terminal.SetContent(0, e.Rows+1, ' ', nil, e.Styles.Status)
	e.Terminal.SetContent(1, e.Rows+1, '', nil, e.Styles.Status)
	e.Terminal.SetContent(2, e.Rows+1, '❯', nil, e.Styles.Status)

	BufferOffset := 3
	for col = BufferOffset; col < e.Cols+e.LineCountWidth; col++ {
		e.Terminal.SetContent(col, e.Rows+1, ' ', nil, e.Styles.Status)
		if col-BufferOffset < len(e.InputBuffer) {
			e.Terminal.SetContent(col, e.Rows+1,
				e.InputBuffer[col-BufferOffset],
				nil, e.Styles.Status)
		}
	}

	var currentLine = e.CursorY + e.OffsetY
	var lineNumberStr = strconv.Itoa(currentLine + 1)
	var currentColumn = e.CursorX + e.OffsetX - e.LineCountWidth
	var columnNumberStr = strconv.Itoa(currentColumn + 1)
	// #TODO do the offsets more neat
	e.PrintMessageStyle(e.Cols, e.Rows+1, e.Styles.Status, columnNumberStr)
	e.PrintMessageStyle(e.Cols-4, e.Rows+1, e.Styles.Status, "col")
	e.PrintMessageStyle(e.Cols-8, e.Rows+1, e.Styles.Status, lineNumberStr)
	e.PrintMessageStyle(e.Cols-12, e.Rows+1, e.Styles.Status, "row")
}

func (e *EditorCore) DisplayLineNumber(row int, textBufferRow int) {
	lineNumberStr := "~"

	if textBufferRow < len(e.TextBuffer) {
		lineNumberStr = strconv.Itoa(textBufferRow + 1)
	}

	lineNumberOffset := e.LineCountWidth - len(lineNumberStr)
	if lineNumberOffset > 0 {
		for i := 0; i < lineNumberOffset; i++ {
			e.Terminal.SetContent(i, row, ' ', nil, e.Styles.Linecount)
		}
	}

	e.PrintMessageStyle(lineNumberOffset, row, e.Styles.Linecount, lineNumberStr)
}

func (e *EditorCore) DisplaySettingsLoop(currentPos int) {
	//Offset between setting names and colors
	colorOffset := "  "
	e.Terminal.SetContent(0, currentPos, ' ', nil, tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack))

	mainFg, mainBg, _ := e.Styles.Main.Decompose()
	e.PrintMessageStyle(1, 0, tcell.StyleDefault.Foreground(mainFg), "Main Foreground"+colorOffset+mainFg.String())
	e.PrintMessageStyle(1, 1, tcell.StyleDefault.Foreground(mainBg), "Main Background"+colorOffset+mainBg.String())

	statusFg, statusBg, _ := e.Styles.Status.Decompose()
	e.PrintMessageStyle(1, 2, tcell.StyleDefault.Foreground(statusFg), "Status Foreground"+colorOffset+statusFg.String())
	e.PrintMessageStyle(1, 3, tcell.StyleDefault.Foreground(statusBg), "Status Background"+colorOffset+statusBg.String())

	messageFg, messageBg, _ := e.Styles.Message.Decompose()
	e.PrintMessageStyle(1, 4, tcell.StyleDefault.Foreground(messageFg), "Message Foreground"+colorOffset+messageFg.String())
	e.PrintMessageStyle(1, 5, tcell.StyleDefault.Foreground(messageBg), "Message Background"+colorOffset+messageBg.String())

	linecountFg, linecountBg, _ := e.Styles.Linecount.Decompose()
	e.PrintMessageStyle(1, 6, tcell.StyleDefault.Foreground(linecountFg), "Linecount Foreground"+colorOffset+linecountFg.String())
	e.PrintMessageStyle(1, 7, tcell.StyleDefault.Foreground(linecountBg), "Linecount Background"+colorOffset+linecountBg.String())

	errorFg, errorBg, _ := e.Styles.Error.Decompose()
	e.PrintMessageStyle(1, 8, tcell.StyleDefault.Foreground(errorFg), "Error Foreground"+colorOffset+errorFg.String())
	e.PrintMessageStyle(1, 9, tcell.StyleDefault.Foreground(errorBg), "Error Background"+colorOffset+errorBg.String())

}

func (e *EditorCore) DisplayColorsLoop(offset int) {
	//LineCount
	e.PrintMessageStyle(0, 10+offset, e.Styles.Linecount, "~1")
	e.PrintMessageStyle(0, 10+1+offset, e.Styles.Linecount, "~2")
	e.PrintMessageStyle(0, 10+2+offset, e.Styles.Linecount, "~3")
	e.PrintMessageStyle(0, 10+3+offset, e.Styles.Linecount, "~4")
	e.PrintMessageStyle(0, 10+4+offset, e.Styles.Linecount, "~5")
	e.PrintMessageStyle(0, 10+5+offset, e.Styles.Linecount, "~6")
	//Main
	e.PrintMessageStyle(2, 10+0+offset, e.Styles.Main, "This is a piece of text! Some characters for testing: ! # ¤ % & / [] {}")

	//Statusbar
	e.PrintMessageStyle(0, 10+6+offset, e.Styles.Status, "write                                                     row 0 col 0")
	//MSG
	e.PrintMessageStyle(30, 10+3+offset, e.Styles.Message, "Open file:")
	e.PrintMessageStyle(30, 10+4+offset, e.Styles.Message, "file.txt  ")
}
