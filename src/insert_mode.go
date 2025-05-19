
package main

func splice(original string, position int, char byte) string {
    byteSlice := []byte(original)

	// Using the null byte as a sentinel value to indicate
	// that the char at the given index should just be removed
    if char == 0 {
        byteSlice = append(byteSlice[:position], byteSlice[position+1:]...)
    } else {
        byteSlice = append(byteSlice[:position], append([]byte{char}, byteSlice[position:]...)...)
    }

    return string(byteSlice)
}


func insertMode[T Terminal](input byte, prog *Program[T]) {
	if input == Escape {
		prog.changeMode(NormalMode)
		return
	}

	panel := prog.getActivePanel()
	buffer := prog.getActiveBuffer()
	line := buffer.lines[panel.logicalCursorY]

	if input == Backspace {
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
		updatedLine := splice(line, panel.logicalCursorX-1, 0)
		buffer.updateLine(panel.logicalCursorY, updatedLine)
		prog.setLogicalCursorPosition(panel.logicalCursorX-1, panel.logicalCursorY)
		return
	}

	updatedLine := splice(line, panel.logicalCursorX, input)
	buffer.lines[panel.logicalCursorY] = updatedLine
	prog.setLogicalCursorPosition(panel.logicalCursorX+1, panel.logicalCursorY)
}

