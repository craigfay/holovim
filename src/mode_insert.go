package main

func splice(original string, position int, r rune) string {
	slice := []rune(original)

	if r == RuneDelete {
		slice = append(slice[:position], slice[position+1:]...)
	} else {
		slice = append(slice[:position], append([]rune{r}, slice[position:]...)...)
	}

	return string(slice)
}

func insertMode[T Terminal](input rune, prog *Program[T]) {
	if input == RuneEscape {
		prog.changeMode(NormalMode)
		return
	}

	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()
	line := buffer.lines[panel.logicalCursorY]

	writeRune := func(r rune) {
		updatedLine := splice(line, panel.logicalCursorX, r)
		buffer.lines[panel.logicalCursorY] = updatedLine
		prog.setLogicalCursorPosition(panel.logicalCursorX+1, panel.logicalCursorY)
	}

	if isStandardUnicode(input) {
		writeRune(input)
		return
	}

	if input == RuneBackspace || input == RuneDelete {
		isFirstLine := panel.logicalCursorY == 0
		isFirstChar := panel.logicalCursorX == 0

		// Doing nothing if at the beginning of the file
		if isFirstChar && isFirstLine {
			return
		}

		// Wrapping the current line back onto the previous line
		if isFirstChar && !isFirstLine {
			prevLine := buffer.lines[panel.logicalCursorY-1]
			newPrevLine := prevLine + line
			buffer.updateLine(panel.logicalCursorY-1, newPrevLine)
			buffer.removeLine(panel.logicalCursorY)
			prog.setLogicalCursorPosition(len(prevLine), panel.logicalCursorY-1)
			return
		}

		// Removing the previous char and updating
		updatedLine := splice(line, panel.logicalCursorX-1, RuneDelete)
		buffer.updateLine(panel.logicalCursorY, updatedLine)
		prog.setLogicalCursorPosition(panel.logicalCursorX-1, panel.logicalCursorY)
		return
	}

	if input == RuneEnter || input == RuneCarriageReturn {
		left := line[:panel.logicalCursorX]
		right := line[panel.logicalCursorX:]

		buffer.updateLine(panel.logicalCursorY, left)
		buffer.insertLine(panel.logicalCursorY+1, right)

		prog.setLogicalCursorPosition(0, panel.logicalCursorY+1)
		return
	}
}
