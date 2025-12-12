package core

import "github.com/gdamore/tcell/v2"

func (e *EditorCore) loopWrite() {
	e.Terminal.Clear()
	e.DisplayBuffer()
	e.DisplayStatus()
	e.Terminal.ShowCursor(e.CursorX, e.CursorY)
	e.Terminal.Show()
	for {
		event := e.Terminal.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
			if mod == tcell.ModNone {
				switch key {
				case tcell.KeyUp:
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
				case tcell.KeyDown:
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
				case tcell.KeyLeft:
					if e.CursorX > e.LineCountWidth {
						e.CursorX--
						// Horizontal scroll left if needed
						if e.CursorX < e.LineCountWidth {
							e.CursorX = e.LineCountWidth
						}
					} else if e.OffsetX > 0 {
						e.OffsetX--
					}
				case tcell.KeyRight:
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
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					// If at the left edge and more to the left, scroll left before deleting
					if e.CursorX == e.LineCountWidth && e.OffsetX > 0 {
						e.OffsetX--
					}
					e.deleteAtCursor()
					// Auto-scroll if cursor goes above visible area
					if e.CursorY < 0 {
						e.OffsetY += e.CursorY
						e.CursorY = 0
					}
					// After deletion, if no characters are visible in the current row, scroll left
					visibleRow := e.CursorY + e.OffsetY
					if visibleRow >= 0 && visibleRow < len(e.TextBuffer) {
						line := e.TextBuffer[visibleRow]
						if e.OffsetX >= len(line) && e.OffsetX > 0 {
							e.OffsetX--
							e.CursorX++
						}
					}
					// Horizontal scroll left if needed after delete
					if e.CursorX < e.LineCountWidth && e.OffsetX > 0 {
						e.OffsetX--
						e.CursorX = e.LineCountWidth
					}
					// If at left edge and more to the left, scroll to show next char to be deleted
					if e.CursorX == e.LineCountWidth && e.OffsetX > 0 {
						e.OffsetX--
					}
				case tcell.KeyEnter:
					e.insertEnter()
					// Auto-scroll if cursor goes below visible area
					if e.CursorY >= e.Rows {
						e.OffsetY += e.CursorY - e.Rows + 1
						e.CursorY = e.Rows - 1
					}
				case tcell.KeyEsc:
					return
				default:
					e.insertRune(ch)
					// Ensure cursor is visible after insertion (horizontal scroll)
					if e.CursorX >= e.Cols+e.LineCountWidth {
						e.OffsetX++
						e.CursorX = e.Cols + e.LineCountWidth - 1
					}
					if e.CursorX < e.LineCountWidth {
						if e.OffsetX > 0 {
							e.OffsetX--
							e.CursorX = e.LineCountWidth
						}
					}
					// Clamp cursor to end of line after insert
					lineLen := len(e.TextBuffer[e.CursorY+e.OffsetY])
					if e.CursorX-e.LineCountWidth+e.OffsetX > lineLen {
						e.CursorX = lineLen - e.OffsetX + e.LineCountWidth
						if e.CursorX < e.LineCountWidth {
							e.CursorX = e.LineCountWidth
						}
					}
				}
			} else if mod == tcell.ModCtrl {
				switch key {
				case tcell.KeyLeft:
					if e.CursorY+e.OffsetY > 0 {
						// Only allow moving right if not past end of line
						if e.CursorX-e.LineCountWidth+e.OffsetX > 0 {
							currChar := 'a'
							// While loop here
							for currChar != ' ' {
								e.CursorX--
								// Horizontal scroll right if needed
								if e.CursorX < e.Cols-e.LineCountWidth {
									e.OffsetX--
									e.CursorX = e.Cols - e.LineCountWidth - 1
								}
								// Check bounds before accessing array
								currentPos := e.CursorX - e.LineCountWidth + e.OffsetX
								if currentPos == 0 {
									currChar = ' '
									break
								}
								currChar = e.TextBuffer[e.CursorY+e.OffsetY][currentPos]
							}
						}
					}
				case tcell.KeyRight:
					if e.CursorY+e.OffsetY < len(e.TextBuffer) {
						// Only allow moving right if not past end of line
						lineLen := len(e.TextBuffer[e.CursorY+e.OffsetY])
						if e.CursorX-e.LineCountWidth+e.OffsetX < lineLen {
							currChar := 'a'
							// While loop here
							for currChar != ' ' {
								e.CursorX++
								// Horizontal scroll right if needed
								if e.CursorX >= e.Cols-e.LineCountWidth {
									e.OffsetX++
									e.CursorX = e.Cols - e.LineCountWidth - 1
								}
								// Check bounds before accessing array
								currentPos := e.CursorX - e.LineCountWidth + e.OffsetX
								if currentPos >= lineLen {
									currChar = ' '
									break
								}
								currChar = e.TextBuffer[e.CursorY+e.OffsetY][currentPos]
							}
						}
					}
				default:
				}
			} else if mod == tcell.ModAlt {

			}

			// Ensure cursor stays within bounds
			if e.CursorY < 0 {
				e.CursorY = 0
			}

			if e.CursorY >= e.Rows {
				e.CursorY = e.Rows - 1
			}

			if e.CursorX < e.LineCountWidth {
				e.CursorX = e.LineCountWidth
			}

			if e.CursorX >= e.Cols+e.LineCountWidth {
				e.CursorX = e.Cols + e.LineCountWidth - 1
			}

			//TODO:termbox.SetCursor(e.CursorX, e.CursorY)
			e.Terminal.Clear()
			e.DisplayBuffer()
			e.DisplayStatus()
			e.Terminal.ShowCursor(e.CursorX, e.CursorY)
			e.Terminal.Show()
		}
	}
}

func (e *EditorCore) insertEnter() {
	CursorPosXinBuffer := e.CursorX - e.LineCountWidth + e.OffsetX
	CursorPosYinBuffer := e.CursorY + e.OffsetY

	if CursorPosYinBuffer < 0 || CursorPosYinBuffer >= len(e.TextBuffer) {
		return
	}

	if CursorPosXinBuffer < 0 {
		CursorPosXinBuffer = 0
	}
	if CursorPosXinBuffer > len(e.TextBuffer[CursorPosYinBuffer]) {
		CursorPosXinBuffer = len(e.TextBuffer[CursorPosYinBuffer])
	}

	currentLine := e.TextBuffer[CursorPosYinBuffer]
	beforeCursor := make([]rune, CursorPosXinBuffer)
	copy(beforeCursor, currentLine[:CursorPosXinBuffer])

	afterCursor := make([]rune, len(currentLine)-CursorPosXinBuffer)
	copy(afterCursor, currentLine[CursorPosXinBuffer:])

	newTextBuffer := make([][]rune, len(e.TextBuffer)+1)

	copy(newTextBuffer[:CursorPosYinBuffer], e.TextBuffer[:CursorPosYinBuffer])

	newTextBuffer[CursorPosYinBuffer] = beforeCursor
	newTextBuffer[CursorPosYinBuffer+1] = afterCursor

	copy(newTextBuffer[CursorPosYinBuffer+2:], e.TextBuffer[CursorPosYinBuffer+1:])
	e.TextBuffer = newTextBuffer
	e.CursorX = e.LineCountWidth
	e.CursorY++

}

func (e *EditorCore) insertRune(insertrune rune) {
	CursorPosXinBuffer := e.CursorX - e.LineCountWidth + e.OffsetX
	CursorPosYinBuffer := e.CursorY + e.OffsetY

	if CursorPosYinBuffer < 0 ||
		CursorPosYinBuffer >= len(e.TextBuffer) ||
		CursorPosXinBuffer < 0 ||
		CursorPosXinBuffer > len(e.TextBuffer[CursorPosYinBuffer]) {
		e.PrintMessageStyle(0, 0, e.Styles.Error, "INSERT WAS NOT INBOUND")
		//termbox.PollEvent()
		return
	}

	line := e.TextBuffer[CursorPosYinBuffer]
	newLine := make([]rune, len(line)+1)
	copy(newLine, line[:CursorPosXinBuffer])
	newLine[CursorPosXinBuffer] = insertrune
	copy(newLine[CursorPosXinBuffer+1:], line[CursorPosXinBuffer:])
	e.TextBuffer[CursorPosYinBuffer] = newLine
	e.CursorX++
}

func (e *EditorCore) deleteAtCursor() {
	CursorPosXinBuffer := e.CursorX - e.LineCountWidth + e.OffsetX
	CursorPosYinBuffer := e.CursorY + e.OffsetY

	if CursorPosYinBuffer < 0 || CursorPosYinBuffer >= len(e.TextBuffer) {
		return
	}

	if CursorPosXinBuffer <= 0 {
		if CursorPosYinBuffer > 0 {
			prevLineLength := len(e.TextBuffer[CursorPosYinBuffer-1])

			e.TextBuffer[CursorPosYinBuffer-1] = append(e.TextBuffer[CursorPosYinBuffer-1], e.TextBuffer[CursorPosYinBuffer]...)

			newTextBuffer := make([][]rune, len(e.TextBuffer)-1)
			copy(newTextBuffer[:CursorPosYinBuffer], e.TextBuffer[:CursorPosYinBuffer])
			copy(newTextBuffer[CursorPosYinBuffer:], e.TextBuffer[CursorPosYinBuffer+1:])
			e.TextBuffer = newTextBuffer
			e.CursorX = prevLineLength + e.LineCountWidth
			e.CursorY--
			return
		}
	} else {
		if CursorPosXinBuffer <= len(e.TextBuffer[CursorPosYinBuffer]) {
			beforeSlice := e.TextBuffer[CursorPosYinBuffer][:CursorPosXinBuffer-1]
			afterSlice := e.TextBuffer[CursorPosYinBuffer][CursorPosXinBuffer:]
			e.TextBuffer[CursorPosYinBuffer] = append(beforeSlice, afterSlice...)
			e.CursorX--
			return
		}
	}
}
