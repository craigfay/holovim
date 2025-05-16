package main

import (
	"bufio"
	"fmt"
	xterm "golang.org/x/term"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	// Extracting the filename from the command-line arguments, and opening it
	filename := args[1]

	file, err := os.Open(filename)

	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return
	}

	// Ensuring the file is closed when the program exits
	defer file.Close()

	// Loading the file contents into a list of strings, line by line
	lines := []string{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file %s\n", err)
		return
	}

	buffers := []Buffer{
		{
			filepath:          filename,
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
			fmt.Println("Error reading input:", err)
			break
		}
		if done {
			break
		}

		handleUserInput(input, prog)

		if prog.state.shouldExit {
			return
		}
	}
}

func handleUserInput[T Terminal](input byte, prog *Program[T]) {
	keys := &prog.settings.keybind

	if input == keys.cursor_down {
		prog.moveCursorDown()
	}

	if input == keys.cursor_up {
		prog.moveCursorUp()
	}

	if input == keys.cursor_left {
		prog.moveCursorLeft()
	}

	if input == keys.cursor_right {
		prog.moveCursorRight()
	}

	if input == keys.close_buffer {
		prog.state.shouldExit = true
	}

	prog.state.needsRedraw = true
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

			line := buffer.lines[lineIdx]

			isActiveLine := lineIdx == panel.logicalCursorY

			// Calculating visual cursor position
			if isActivePanel && isActiveLine {
				x := countVisualColumns(line, panel.logicalCursorX, &prog.settings)
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

func getLogger(filename string) func(string) error {
	// Getting the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get current directory: %v", err))
	}

	// Constructing the full path to the file in the current directory
	fullPath := filepath.Join(currentDir, filename)

	// Ensuring the logfile is empty by truncating or creating it
	file, err := os.Create(fullPath)
	if err != nil {
		panic(fmt.Sprintf("failed to create or truncate file: %v", err))
	}
	file.Close() // Closing the file immediately after truncating it

	// Returning a function that appends strings to the file
	return func(logMessage string) error {
		// Opening the file in append mode, creating it if it doesn't exist
		file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open or create file: %w", err)
		}
		defer file.Close()

		// Writing the log message to the file
		_, err = file.WriteString(logMessage + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}

		return nil
	}
}
