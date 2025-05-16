package main

func getVisualX(line string, logicalX int, settings *Settings) int {
	result := 0
	for i := 0; i < min(logicalX, len(line)); i++ {
		if line[i] == '\t' {
			result += settings.tabstop
		} else {
			result += 1
		}
	}
	return result
}

func getLogicalXWithVisualX(line string, visualX int, settings *Settings) int {
	newLogicalX := 0
	newVisualX := 0

	// Incrementing newLogicalX until another increment would
	// exceed the previous visualCursorX
	for {
		if newLogicalX+1 >= len(line) {
			break
		}

		if newVisualX >= visualX {
			break
		}

		visualXChunk := 0

		isTab := line[newLogicalX] == '\t'

		if isTab {
			visualXChunk += settings.tabstop
		} else {
			visualXChunk += 1
		}

		if newVisualX+visualXChunk > visualX {
			break
		}

		newVisualX += visualXChunk
		newLogicalX += 1
	}

	return newLogicalX
}
