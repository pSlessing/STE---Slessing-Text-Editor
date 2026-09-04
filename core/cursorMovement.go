package core

// Shared cursor-movement logic used by both command mode (editor.go's
// inputHandling) and write mode (write.go's loopWrite). Previously this
// logic was copy-pasted between the two, which let them drift out of sync.

// clampCursorXToLine pulls CursorX back onto the current line when a
// vertical move lands on a shorter line than the one the cursor came from.
func (e *EditorCore) clampCursorXToLine() {
	if e.CursorY+e.OffsetY < len(e.TextBuffer) && e.CursorX-e.LineCountWidth > len(e.TextBuffer[e.CursorY+e.OffsetY]) {
		e.CursorX = len(e.TextBuffer[e.CursorY+e.OffsetY]) + e.LineCountWidth
	}
}

func (e *EditorCore) moveCursorUp() {
	if e.CursorY > 0 {
		// Move cursor up within visible area
		e.CursorY--
	} else if e.OffsetY > 0 {
		// Scroll up when cursor is at top
		e.OffsetY--
	}
	e.clampCursorXToLine()
}

func (e *EditorCore) moveCursorDown() {
	if e.CursorY < e.Rows-1 && e.CursorY+e.OffsetY+1 < len(e.TextBuffer) {
		// Move cursor down within visible area
		e.CursorY++
	} else if e.OffsetY+e.Rows < len(e.TextBuffer) {
		// Scroll down when cursor is at bottom
		e.OffsetY++
	}
	e.clampCursorXToLine()
}

// moveCursorUpSkipEmpty moves up at least one line, then keeps moving up
// past any blank lines (used for Ctrl+Up "paragraph" jumps).
func (e *EditorCore) moveCursorUpSkipEmpty() {
	e.moveCursorUp()
	currentLine := e.TextBuffer[e.CursorY+e.OffsetY]
	for len(currentLine) == 0 {
		e.moveCursorUp()
		currentLine = e.TextBuffer[e.CursorY+e.OffsetY]
	}
}

// moveCursorDownSkipEmpty is the Ctrl+Down counterpart of moveCursorUpSkipEmpty.
func (e *EditorCore) moveCursorDownSkipEmpty() {
	e.moveCursorDown()
	currentLine := e.TextBuffer[e.CursorY+e.OffsetY]
	for len(currentLine) == 0 {
		e.moveCursorDown()
		currentLine = e.TextBuffer[e.CursorY+e.OffsetY]
	}
}

func (e *EditorCore) moveCursorLeft() {
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

func (e *EditorCore) moveCursorRight() {
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

// moveCursorWordLeft implements Ctrl+Left: jump to the start of the
// previous word on the current line.
func (e *EditorCore) moveCursorWordLeft() {
	if e.CursorX-e.LineCountWidth+e.OffsetX > 0 {
		currChar := 'a'
		for currChar != ' ' {
			e.CursorX--
			// Horizontal scroll left if needed
			if e.CursorX < 0 {
				e.OffsetX--
				e.CursorX = 0
			}
			currentPos := e.CursorX - e.LineCountWidth + e.OffsetX
			if currentPos == 0 {
				currChar = ' '
				break
			}
			currChar = e.TextBuffer[e.CursorY+e.OffsetY][currentPos]
		}
	}
}

// moveCursorWordRight is the Ctrl+Right counterpart of moveCursorWordLeft.
func (e *EditorCore) moveCursorWordRight() {
	if e.CursorY+e.OffsetY < len(e.TextBuffer) {
		lineLen := len(e.TextBuffer[e.CursorY+e.OffsetY])
		if e.CursorX-e.LineCountWidth+e.OffsetX < lineLen {
			currChar := 'a'
			for currChar != ' ' {
				e.CursorX++
				// Horizontal scroll right if needed
				if e.CursorX >= e.Cols-e.LineCountWidth {
					e.OffsetX++
					e.CursorX = e.Cols - e.LineCountWidth - 1
				}
				currentPos := e.CursorX - e.LineCountWidth + e.OffsetX
				if currentPos >= lineLen {
					currChar = ' '
					break
				}
				currChar = e.TextBuffer[e.CursorY+e.OffsetY][currentPos]
			}
		}
	}
}
