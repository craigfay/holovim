package main

func (p *Program[T]) moveCursorDown() {
	settings := &p.settings
	s := &p.state

	panel := &s.panels[s.activePanelIdx]
	buffer := &s.buffers[panel.bufferIdx]

	isAtContentBottom := panel.logicalCursorY+1 >= len(buffer.lines)

	canScroll := buffer.topVisibleLineIdx+panel.height+1 < len(buffer.lines)
	isAtViewportBottom := s.visualCursorY == panel.topLeftY+panel.height

	if !isAtContentBottom || canScroll {
		// moving the cursor down
		nextLine := buffer.lines[panel.logicalCursorY+1]
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

		p.setVisualCursorPosition(newVisualX, newVisualY)
		p.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY+1)
	}
}

func (p *Program[T]) moveCursorUp() {
	settings := p.settings
	s := p.state

	panel := &s.panels[s.activePanelIdx]
	buffer := &s.buffers[panel.bufferIdx]

	canScroll := buffer.topVisibleLineIdx > 0

	if panel.logicalCursorY > 0 || canScroll {
		prevLine := buffer.lines[panel.logicalCursorY-1]
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

		p.setVisualCursorPosition(newVisualX, newVisualY)
		p.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY-1)
	}
}

func (p *Program[T]) moveCursorLeft() {
	settings := p.settings
	s := p.state

	panel := &s.panels[s.activePanelIdx]
	buffer := &s.buffers[panel.bufferIdx]

	lineContent := &buffer.lines[panel.logicalCursorY]

	// Wrapping to the end of the previous line
	if panel.logicalCursorX == 0 && panel.logicalCursorY != 0 {
		if !settings.cursor_x_overflow {
			return
		}
		prevLine := buffer.lines[panel.logicalCursorY-1]
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

		p.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY-1)
		p.setVisualCursorPosition(newVisualX, newVisualY)
		s.bookmarkedVisualCursorX = newVisualX

	} else if panel.logicalCursorX != 0 {
		// Moving the cursor left within the current line
		thisChar := (*lineContent)[panel.logicalCursorX-1]
		newVisualX := s.visualCursorX

		if thisChar == '\t' {
			newVisualX -= settings.tabstop
		} else {
			newVisualX -= 1
		}

		p.setLogicalCursorPosition(panel.logicalCursorX-1, panel.logicalCursorY)
		p.setVisualCursorPosition(newVisualX, s.visualCursorY)
		s.bookmarkedVisualCursorX = newVisualX
	}
}

func (p *Program[T]) moveCursorRight() {
	settings := p.settings
	s := p.state

	panel := &s.panels[s.activePanelIdx]
	buffer := &s.buffers[panel.bufferIdx]

	lineContent := &buffer.lines[panel.logicalCursorY]
	lineLength := len(*lineContent)
	isAtEndOfLine := panel.logicalCursorX+1 >= lineLength
	isLastLine := panel.logicalCursorY == len(buffer.lines)-1
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

		p.setLogicalCursorPosition(0, panel.logicalCursorY+1)
		p.setVisualCursorPosition(s.leftChromeWidth, newVisualY)
		s.bookmarkedVisualCursorX = s.leftChromeWidth

		// moving the cursor right
	} else {
		thisChar := (*lineContent)[panel.logicalCursorX]
		newVisualX := s.visualCursorX

		if thisChar == '\t' {
			newVisualX += settings.tabstop
		} else {
			newVisualX += 1
		}

		p.setLogicalCursorPosition(panel.logicalCursorX+1, panel.logicalCursorY)
		p.setVisualCursorPosition(newVisualX, s.visualCursorY)
		s.bookmarkedVisualCursorX = newVisualX
	}
}
