package main

func (prog *Program[T]) moveCursorDown() {
	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()
	isAtContentBottom := panel.logicalCursorY+1 >= len(buffer.lines)
	canScroll := buffer.topVisibleLineIdx+panel.height+1 < len(buffer.lines)
	isAtViewportBottom := prog.state.visualCursorY == panel.topLeftY+panel.height

	if !isAtContentBottom || canScroll {
		// Moving the cursor down
		line := buffer.lineContent(panel.logicalCursorY)
		nextLine := buffer.lineContent(panel.logicalCursorY + 1)

		currentVisualX := getVisualX(line, panel.logicalCursorX, &prog.settings)
		newLogicalX := getLogicalXWithVisualX(nextLine, currentVisualX, &prog.settings)

		// Respecting pinned visual x
		if currentVisualX < panel.pinnedVisualCursorX {
			newLogicalX = getLogicalXWithVisualX(nextLine, panel.pinnedVisualCursorX, &prog.settings)
		}

		// Scrolling if necessary
		if isAtViewportBottom {
			buffer.topVisibleLineIdx += 1
		}

		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY+1)
	}
}

func (prog *Program[T]) moveCursorUp() {
	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()
	canScroll := buffer.topVisibleLineIdx > 0

	if panel.logicalCursorY > 0 || canScroll {
		line := buffer.lineContent(panel.logicalCursorY)
		prevLine := buffer.lineContent(panel.logicalCursorY - 1)

		currentVisualX := getVisualX(line, panel.logicalCursorX, &prog.settings)
		newLogicalX := getLogicalXWithVisualX(prevLine, currentVisualX, &prog.settings)

		// Scrolling if necessary
		if prog.state.visualCursorY == prog.state.topChromeHeight {
			buffer.topVisibleLineIdx -= 1
		}

		// Respecting pinned visual x
		if currentVisualX < panel.pinnedVisualCursorX {
			newLogicalX = getLogicalXWithVisualX(prevLine, panel.pinnedVisualCursorX, &prog.settings)
		}

		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY-1)
	}
}

func (prog *Program[T]) moveCursorLeft() {
	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()

	// Wrapping to the end of the previous line
	if panel.logicalCursorX == 0 && panel.logicalCursorY != 0 {
		if !prog.settings.cursor_x_overflow {
			return
		}
		prevLine := buffer.lineContent(panel.logicalCursorY - 1)
		newLogicalX := max(len(prevLine)-1, 0)

		// Scrolling if necessary
		if prog.state.visualCursorY == prog.state.topChromeHeight {
			buffer.topVisibleLineIdx -= 1
		}

		panel.pinnedVisualCursorX = getVisualX(prevLine, newLogicalX, &prog.settings)
		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY-1)

	} else if panel.logicalCursorX != 0 {
		// Moving the cursor left within the current line
		newLogicalX := panel.logicalCursorX - 1
		line := buffer.lineContent(panel.logicalCursorY)
		panel.pinnedVisualCursorX = getVisualX(line, newLogicalX, &prog.settings)
		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY)
	}
}

func (prog *Program[T]) moveCursorRight() {
	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()

	line := buffer.lineContent(panel.logicalCursorY)
	lineLength := len(line)
	isAtEndOfLine := panel.logicalCursorX+1 >= lineLength
	isLastLine := panel.logicalCursorY == len(buffer.lines)-1
	isAtViewportBottom := prog.state.visualCursorY == panel.topLeftY+panel.height

	if isAtEndOfLine && isLastLine {
		return
	}

	if isAtEndOfLine && !isLastLine {
		// wrapping to the beginning of the next line
		if !prog.settings.cursor_x_overflow {
			return
		}

		// scrolling if necessary
		if isAtViewportBottom {
			buffer.topVisibleLineIdx += 1
		}

		panel.pinnedVisualCursorX = 0
		prog.setLogicalCursorPosition(0, panel.logicalCursorY+1)

	} else {
		// moving the cursor right
		newLogicalX := panel.logicalCursorX + 1
		panel.pinnedVisualCursorX = getVisualX(line, newLogicalX, &prog.settings)
		prog.setLogicalCursorPosition(newLogicalX, panel.logicalCursorY)
	}
}
