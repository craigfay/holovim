package main

func (prog *Program[T]) moveCursorDown() {
	s := &prog.state
	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()
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
	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()
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

func (prog *Program[T]) moveCursorLeft() {
	settings := prog.settings
	s := prog.state

	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()

	// Wrapping to the end of the previous line
	if panel.logicalCursorX == 0 && panel.logicalCursorY != 0 {
		if !settings.cursor_x_overflow {
			return
		}
		prevLine := buffer.lines[panel.logicalCursorY-1]
		newLogicalX := max(len(prevLine)-1, 0)

		// Scrolling if necessary
		if s.visualCursorY == s.topChromeHeight {
			buffer.topVisibleLineIdx -= 1
		}

		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY-1)

	} else if panel.logicalCursorX != 0 {
		// Moving the cursor left within the current line
		prog.setLogicalCursorPosition(panel.logicalCursorX-1, panel.logicalCursorY)
	}
}

func (prog *Program[T]) moveCursorRight() {
	settings := prog.settings
	s := prog.state

	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()

	lineContent := &buffer.lines[panel.logicalCursorY]
	lineLength := len(*lineContent)
	isAtEndOfLine := panel.logicalCursorX+1 >= lineLength
	isLastLine := panel.logicalCursorY == len(buffer.lines)-1
	isAtViewportBottom := s.visualCursorY == panel.topLeftY+panel.height

	if isAtEndOfLine && isLastLine {
		return
	}

	// wrapping to the beginning of the next line
	if isAtEndOfLine && !isLastLine {
		if !settings.cursor_x_overflow {
			return
		}

		// scrolling if necessary
		if isAtViewportBottom {
			buffer.topVisibleLineIdx += 1
		}

		prog.setLogicalCursorPosition(0, panel.logicalCursorY+1)


	} else {
		// moving the cursor right
		prog.setLogicalCursorPosition(panel.logicalCursorX+1, panel.logicalCursorY)

	}
}
