package main

func (prog *Program[T]) moveCursorDown() {
	s := &prog.state
	panel := prog.getActivePanel()
	buffer := &s.buffers[panel.bufferIdx]
	isAtContentBottom := panel.logicalCursorY+1 >= len(buffer.lines)
	canScroll := buffer.topVisibleLineIdx+panel.height+1 < len(buffer.lines)
	isAtViewportBottom := s.visualCursorY == panel.topLeftY+panel.height

	if !isAtContentBottom || canScroll {
		// Moving the cursor down
		line := buffer.lines[panel.logicalCursorY]
		nextLine := buffer.lines[panel.logicalCursorY+1]

		currentVisualX := getVisualX(line, panel.logicalCursorX, &prog.settings)
		newLogicalX := getLogicalXWithVisualX(nextLine, currentVisualX, &prog.settings)

		// Scrolling if necessary
		if isAtViewportBottom {
			buffer.topVisibleLineIdx += 1
		}

		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY+1)
	}
}

func (prog *Program[T]) moveCursorUp() {
	s := prog.state

	tab := &s.tabs[s.activeTabIdx]
	panel := &tab.panels[tab.activePanelIdx]
	buffer := &s.buffers[panel.bufferIdx]

	canScroll := buffer.topVisibleLineIdx > 0

	if panel.logicalCursorY > 0 || canScroll {
		line := buffer.lines[panel.logicalCursorY]
		prevLine := buffer.lines[panel.logicalCursorY-1]

		currentVisualX := getVisualX(line, panel.logicalCursorX, &prog.settings)
		newLogicalX := getLogicalXWithVisualX(prevLine, currentVisualX, &prog.settings)

		// Scrolling if necessary
		if s.visualCursorY == s.topChromeHeight {
			buffer.topVisibleLineIdx -= 1
		}

		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY-1)
	}
}

func (p *Program[T]) moveCursorLeft() {
	settings := p.settings
	s := p.state

	tab := &s.tabs[s.activeTabIdx]
	panel := &tab.panels[tab.activePanelIdx]
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

	tab := &s.tabs[s.activeTabIdx]
	panel := &tab.panels[tab.activePanelIdx]
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

	} else {
		// moving the cursor right
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
