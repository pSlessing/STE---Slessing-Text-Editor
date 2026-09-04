package core

import (
	"strconv"
	"strings"
	"time"

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

// cursorScreenX returns the screen X position of the cursor, accounting for tabs.
func (e *EditorCore) cursorScreenX() int {
	bufRow := e.CursorY + e.OffsetY
	bufCol := e.CursorX - e.LineCountWidth + e.OffsetX
	scrollVisual := e.bufferToVisual(bufRow, e.OffsetX)
	cursorVisual := e.bufferToVisual(bufRow, bufCol)
	return cursorVisual - scrollVisual + e.LineCountWidth
}

// bufferToVisual returns the visual column for a given buffer column on a row,
// accounting for tab expansion. bufCol is a buffer index (not screen position).
func (e *EditorCore) bufferToVisual(bufRow, bufCol int) int {
	if bufRow < 0 || bufRow >= len(e.TextBuffer) {
		return bufCol
	}
	visual := 0
	for i, ch := range e.TextBuffer[bufRow] {
		if i >= bufCol {
			break
		}
		if ch == '\t' {
			visual += e.TabSize - (visual % e.TabSize)
		} else {
			visual += runewidth.RuneWidth(ch)
		}
	}
	return visual
}

func (e *EditorCore) DisplayBuffer() {
	for row := 0; row <= e.Rows; row++ {
		textBufferRow := row + e.OffsetY

		e.DisplayLineNumber(row, textBufferRow)

		if textBufferRow < 0 || textBufferRow >= len(e.TextBuffer) {
			continue
		}

		line := e.TextBuffer[textBufferRow]
		// OffsetX is a buffer index — compute its visual column as the scroll origin.
		scrollVisual := e.bufferToVisual(textBufferRow, e.OffsetX)

		for bufCol := e.OffsetX; bufCol < len(line); bufCol++ {
			ch := line[bufCol]
			absVisual := e.bufferToVisual(textBufferRow, bufCol)
			screenCol := absVisual - scrollVisual

			var cellWidth int
			if ch == '\t' {
				cellWidth = e.TabSize - (absVisual % e.TabSize)
			} else {
				cellWidth = runewidth.RuneWidth(ch)
			}

			for i := 0; i < cellWidth; i++ {
				sc := screenCol + i
				if sc >= e.Cols {
					break
				}
				var r rune
				if ch == '\t' || i > 0 {
					r = ' '
				} else {
					r = ch
				}
				e.Terminal.SetContent(sc+e.LineCountWidth, row, r, nil, e.Styles.Main)
			}

			if screenCol+cellWidth >= e.Cols {
				break
			}
		}
	}
}

func (e *EditorCore) DisplayStatus() {
	var col int

	e.Terminal.SetContent(0, e.Rows+1, ' ', nil, e.Styles.Status)
	e.Terminal.SetContent(1, e.Rows+1, '', nil, e.Styles.Status)
	e.Terminal.SetContent(2, e.Rows+1, '❯', nil, e.Styles.Status)

	// statusWidth is the full terminal width in columns; e.Cols excludes the
	// line-count gutter, so the status bar (which spans the whole row) needs
	// to add it back to find the true right edge.
	statusWidth := e.Cols + e.LineCountWidth

	BufferOffset := 3
	for col = BufferOffset; col < statusWidth; col++ {
		e.Terminal.SetContent(col, e.Rows+1, ' ', nil, e.Styles.Status)
		if col-BufferOffset < len(e.InputBuffer) {
			e.Terminal.SetContent(col, e.Rows+1,
				e.InputBuffer[col-BufferOffset],
				nil, e.Styles.Status)
		}
	}

	var currentLine = e.CursorY + e.OffsetY
	var currentColumn = e.CursorX + e.OffsetX - e.LineCountWidth

	rightFields := []string{"row " + strconv.Itoa(currentLine+1), "col " + strconv.Itoa(currentColumn+1)}
	if e.CurrentBranch != "" {
		rightFields = append([]string{"[" + e.CurrentBranch + "]"}, rightFields...)
	}
	rightText := strings.Join(rightFields, "  ")

	// Right-align against the real edge of the status bar instead of
	// hardcoded offsets, so the layout adapts to however wide the row/col
	// numbers or branch name actually are, and to narrow terminals.
	rightStart := statusWidth - len(rightText)
	if rightStart < BufferOffset {
		rightStart = BufferOffset
	}
	e.PrintMessageStyle(rightStart, e.Rows+1, e.Styles.Status, rightText)

	if e.statusMessage != "" && time.Since(e.statusMessageTime) < 3*time.Second {
		lines := strings.Split(e.statusMessage, "\n")
		maxLen := 0
		for _, l := range lines {
			if len(l) > maxLen {
				maxLen = len(l)
			}
		}
		startCol := e.Cols - maxLen
		if startCol < 0 {
			startCol = 0
		}
		// Stack the message directly above the status bar (e.Rows+1) instead
		// of at row 0, so it appears as an overlay near the status bar rather
		// than overwriting the top of the buffer.
		startRow := e.Rows + 1 - len(lines)
		if startRow < 0 {
			startRow = 0
		}
		for i, l := range lines {
			row := startRow + i
			for col := startCol; col <= e.Cols; col++ {
				e.Terminal.SetContent(col, row, ' ', nil, e.Styles.Status)
			}
			e.PrintMessageStyle(startCol, row, e.Styles.Status, l)
		}
	} else {
		e.statusMessage = ""
	}
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
