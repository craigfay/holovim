
package main

func moveCursorDown(s *ProgramState, settings *Settings) {
	buffer := &s.buffers[s.activeBufferIdx]
	contentAreaMaxY := s.termHeight - s.bottomChromeHeight
	contentAreaRowCount := s.termHeight - s.topChromeHeight - s.bottomChromeHeight
	isAtContentBottom := s.logicalCursorY+1 >= len(buffer.lines)
	canScroll := buffer.topVisibleLineIdx+contentAreaRowCount+1 < len(buffer.lines)
	isAtViewportBottom := s.visualCursorY == contentAreaMaxY

	if !isAtContentBottom || canScroll {
		// moving the cursor down
		nextLine := buffer.lines[s.logicalCursorY+1]
		newLogicalX := 0
		newVisualX := s.leftChromeWidth

		targetVisualCursorX := max(s.visualCursorX, s.bookmarkedVisualCursorX)

		// Incrementing newLogicalX until another increment would
		// exceed the previous visualCursorX
		for {
			if newLogicalX+1 >= len(nextLine) {
				break
			}

			if newVisualX >= targetVisualCursorX {
				break
			}

			visualXChunk := 0

			isTab := nextLine[newLogicalX] == '\t'

			if isTab {
				visualXChunk += settings.tabstop
			} else {
				visualXChunk += 1
			}

			if newVisualX+visualXChunk > targetVisualCursorX {
				break
			}

			newVisualX += visualXChunk
			newLogicalX += 1
		}

		newVisualY := s.visualCursorY + 1

		// Scrolling if necessary
		if isAtViewportBottom {
			buffer.topVisibleLineIdx += 1
			newVisualY = s.visualCursorY
		}

		s.setVisualCursorPosition(newVisualX, newVisualY)
		s.setLogicalCursorPosition(newLogicalX, s.logicalCursorY+1)
	}
}

func moveCursorUp(s *ProgramState, settings *Settings) {
	buffer := &s.buffers[s.activeBufferIdx]
	canScroll := buffer.topVisibleLineIdx > 0

	if s.logicalCursorY > 0 || canScroll {
		prevLine := buffer.lines[s.logicalCursorY-1]
		newLogicalX := 0
		newVisualX := s.leftChromeWidth

		targetVisualCursorX := max(s.visualCursorX, s.bookmarkedVisualCursorX)

		for {
			if newLogicalX+1 >= len(prevLine) {
				break
			}

			if newVisualX >= targetVisualCursorX {
				break
			}

			visualXChunk := 0
			isTab := prevLine[newLogicalX] == '\t'

			if isTab {
				visualXChunk += settings.tabstop
			} else {
				visualXChunk += 1
			}

			if newVisualX+visualXChunk > targetVisualCursorX {
				break
			}

			newVisualX += visualXChunk
			newLogicalX += 1
		}

		newVisualY := s.visualCursorY - 1

		if s.visualCursorY == s.topChromeHeight {
			buffer.topVisibleLineIdx -= 1
			newVisualY = s.visualCursorY
		}

		s.setVisualCursorPosition(newVisualX, newVisualY)
		s.setLogicalCursorPosition(newLogicalX, s.logicalCursorY-1)
	}
}

func moveCursorLeft(s *ProgramState, settings *Settings) {
	buffer := &s.buffers[s.activeBufferIdx]
	lineContent := &buffer.lines[s.logicalCursorY]

	// Wrapping to the end of the previous line
	if s.logicalCursorX == 0 && s.logicalCursorY != 0 {
		if !settings.cursor_x_overflow {
			return
		}
		prevLine := buffer.lines[s.logicalCursorY-1]
		newLogicalX := max(len(prevLine)-1, 0)
		newVisualX := s.leftChromeWidth

		// Counting the visual columns in the previous line
		for i := 0; i < len(prevLine)-1; i++ {
			if prevLine[i] == '\t' {
				newVisualX += settings.tabstop
			} else {
				newVisualX += 1
			}
		}

		newVisualY := s.visualCursorY - 1

		// Scrolling if necessary
		if s.visualCursorY == s.topChromeHeight {
			buffer.topVisibleLineIdx -= 1
			newVisualY = s.visualCursorY
		}

		s.setLogicalCursorPosition(newLogicalX, s.logicalCursorY-1)
		s.setVisualCursorPosition(newVisualX, newVisualY)
		s.bookmarkedVisualCursorX = newVisualX

	} else if s.logicalCursorX != 0 {
		// Moving the cursor left within the current line
		thisChar := (*lineContent)[s.logicalCursorX-1]
		newVisualX := s.visualCursorX

		if thisChar == '\t' {
			newVisualX -= settings.tabstop
		} else {
			newVisualX -= 1
		}

		s.setLogicalCursorPosition(s.logicalCursorX-1, s.logicalCursorY)
		s.setVisualCursorPosition(newVisualX, s.visualCursorY)
		s.bookmarkedVisualCursorX = newVisualX
	}
}

func moveCursorRight(s *ProgramState, settings *Settings) {
	buffer := &s.buffers[s.activeBufferIdx]
	lineContent := &buffer.lines[s.logicalCursorY]
	lineLength := len(*lineContent)
	isAtEndOfLine := s.logicalCursorX+1 >= lineLength
	isLastLine := s.logicalCursorY == len(buffer.lines)-1
	contentAreaMaxY := s.termHeight - s.bottomChromeHeight
	isAtViewportBottom := s.visualCursorY == contentAreaMaxY

	if isAtEndOfLine && isLastLine {
		return
	}

	// wrapping to the beginning of the next line
	if isAtEndOfLine && !isLastLine {
		if !settings.cursor_x_overflow {
			return
		}

		newVisualY := s.visualCursorY + 1

		// scrolling if necessary
		if isAtViewportBottom {
			newVisualY = s.visualCursorY
			buffer.topVisibleLineIdx += 1
		}

		s.setLogicalCursorPosition(0, s.logicalCursorY+1)
		s.setVisualCursorPosition(s.leftChromeWidth, newVisualY)
		s.bookmarkedVisualCursorX = s.leftChromeWidth

		// moving the cursor right
	} else {
		thisChar := (*lineContent)[s.logicalCursorX]
		newVisualX := s.visualCursorX

		if thisChar == '\t' {
			newVisualX += settings.tabstop
		} else {
			newVisualX += 1
		}

		s.setLogicalCursorPosition(s.logicalCursorX+1, s.logicalCursorY)
		s.setVisualCursorPosition(newVisualX, s.visualCursorY)
		s.bookmarkedVisualCursorX = newVisualX
	}
}
