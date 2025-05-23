package main

import (
	"bufio"
	"fmt"
	xterm "golang.org/x/term"
	"os"
	"strings"
)

func main() {
	args := os.Args

	// Extracting the filename from the command-line arguments, and opening it
	filepath := args[1]

	status := checkPath(filepath)

	if status == FileStatusNotExists {
		fmt.Printf("File or directory does not exist: %v", filepath)
		return
	}

	if status == FileStatusAccessDenied {
		fmt.Printf("Access denied: %v", filepath)
		return
	}

	if status == FileStatusIsDirectory {
		// TODO open file explorer
		fmt.Printf("Cannot open directories yet: %v", filepath)
		return
	}

	file, err := os.Open(filepath)

	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filepath, err)
		return
	}

	// Ensuring the file is closed when the program exits
	defer file.Close()

	// Loading the file contents into a list of strings, line by line
	lines := []BufferLine{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, BufferLine{
			content: scanner.Text(),
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file %s\n", err)
		return
	}

	buffers := []Buffer{
		{
			filepath:          filepath,
			lines:             lines,
			topVisibleLineIdx: 0,
		},
	}

	program := Program[ANSI]{
		logger:   getLogger("./logfile.log.txt"),
		state:    ProgramState{},
		term:     ANSI{},
		settings: defaultSettings(),
	}

	program.state.buffers = buffers
	program.setCWD(filepath)

	// Saving the current state of the terminal,
	// and re-loading it when this program exits
	oldTerminalState, err := xterm.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		panic(err)
	}

	// Resetting cursor position and terminal state after the program closes.
	// This isn't exactly true, because we're not resetting it exactly as it was.
	defer program.term.setCursorPosition(0, 0)
	defer xterm.Restore(int(os.Stdin.Fd()), oldTerminalState)

	initializeState(&program)

	inputIterator := NewStdinIterator()

	runMainLoop(&program, inputIterator)
}

func runMainLoop[T Terminal](prog *Program[T], inputIterator InputIterator) {
	for {
		prog.updateTopChrome()

		if true || prog.state.needsRedraw {
			redraw(prog)
		}

		done, input, err := inputIterator.Next()
		if err != nil {
			prog.logger(fmt.Sprintf("Error reading input: %v", err))
			break
		}
		if done {
			break
		}

		if prog.state.currentMode == NormalMode {
			normalMode(input, prog)
		} else if prog.state.currentMode == InsertMode {
			insertMode(input, prog)
		}

		if prog.state.shouldExit {
			return
		}
	}
}

func redraw[T Terminal](prog *Program[T]) {
	s := &prog.state
	settings := &prog.settings

	prog.term.clearScreen()
	prog.setVisualCursorPosition(0, 0)

	// Printing top chrome content
	for i := 0; i < s.topChromeHeight; i++ {
		line := s.topChromeContent[i]
		prog.term.printf("%s", line)
		prog.setVisualCursorPosition(s.leftChromeWidth, s.visualCursorY+1)
	}

	tab := prog.state.tabs[prog.state.activeTabIdx]
	panel := &tab.panels[tab.activePanelIdx]
	prog.setVisualCursorPosition(panel.topLeftX, panel.topLeftY)
	buffer := &s.buffers[panel.bufferIdx]

	visualCursorX := 0
	visualCursorY := 0

	for idx, panel := range tab.panels {
		isActivePanel := idx == tab.activePanelIdx

		prog.setVisualCursorPosition(panel.topLeftX, panel.topLeftY)

		// Drawing an individual panel
		for y := 0; y <= panel.height; y++ {

			lineIdx := y + buffer.topVisibleLineIdx

			// Stopping if about to try to draw a line that doesn't exist
			if lineIdx >= len(buffer.lines) {
				break
			}

			line := buffer.lines[lineIdx].content

			isActiveLine := lineIdx == panel.logicalCursorY

			// Calculating visual cursor position
			if isActivePanel && isActiveLine {
				x := getVisualX(line, panel.logicalCursorX, &prog.settings)
				visualCursorY = panel.topLeftY + y
				visualCursorX = panel.topLeftX + x
			}

			// Doing whitespace-related formatting, and printing the current line
			line = replaceTabsWithSpaces(line, settings.tabstop, settings.tabchar)
			lastCharIdx := min(panel.width, len(line))
			prog.term.printf("%s", line[:lastCharIdx])

			prog.setVisualCursorPosition(panel.topLeftX, s.visualCursorY+1)
		}
	}

	prog.setVisualCursorPosition(visualCursorX, visualCursorY)
	s.needsRedraw = false
}

func replaceTabsWithSpaces(line string, tabWidth int, tabchar string) string {
	tabcharLength := len([]rune(tabchar))

	assert(
		tabcharLength == 1,
		fmt.Sprintf(
			"Expected `tabchar` to have length of 1. Actual: %d",
			tabcharLength,
		),
	)

	whitespace := tabchar

	for i := 0; i < tabWidth-1; i++ {
		whitespace += " "
	}

	return strings.ReplaceAll(line, "\t", whitespace)
}

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
