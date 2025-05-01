package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	settings := Settings{
		tabstop:           4,
		tabchar:           "›", // (U+203A)
		cursor_x_overflow: true,
	}

	// Saving the current state of the terminal,
	// and re-loading it when this program exits
	oldTerminalState, err := term.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		panic(err)
	}

	// Resetting cursor position and terminal state after the program closes.
	// This isn't exactly true, because we're not resetting it exactly as it was.
	defer ANSI{}.setCursorPosition(0, 0)
	defer term.Restore(int(os.Stdin.Fd()), oldTerminalState)

	s := initializeState(&settings)

	// Declaring a buffer to store a single byte of user input at a time
	buf := make([]byte, 1)

	for {
		if true || s.needsRedraw {
			redraw(&s, &settings)
		}

		// Reading a single byte from stdin into the buffer
		_, err := os.Stdin.Read(buf)

		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		handleUserInput(buf[0], &s, &settings)

		if s.shouldExit {
			return
		}
	}
}

func handleUserInput(input byte, s *ProgramState, settings *Settings) {
	if input == s.motions.cursor_down {
		moveCursorDown(s, settings)
	}

	if input == s.motions.cursor_up {
		moveCursorUp(s, settings)
	}

	if input == s.motions.cursor_left {
		moveCursorLeft(s, settings)
	}

	if input == s.motions.cursor_right {
		moveCursorRight(s, settings)
	}

	// Exiting the loop if 'q' is pressed
	if input == 'q' {
		s.shouldExit = true
	}

	s.needsRedraw = true
}

func redraw(s *ProgramState, settings *Settings) {
	contentAreaRowCount := s.termHeight - s.topChromeHeight - s.bottomChromeHeight
	buffer := &s.buffers[s.activeBufferIdx]

	preDrawCursorX := s.visualCursorX
	preDrawCursorY := s.visualCursorY

	ANSI{}.clearScreen()
	s.setVisualCursorPosition(0, 0)

	// Printing top chrome content
	for i := 0; i < s.topChromeHeight; i++ {
		line := s.topChromeContent[i]

		fmt.Printf("%s", line)
		s.setVisualCursorPosition(s.leftChromeWidth, s.visualCursorY+1)
	}

	// Printing main buffer content
	for i := 0; i <= contentAreaRowCount; i++ {
		lineIdx := i + buffer.topVisibleLineIdx

		// Stopping if about to try to draw a line that doesn't exist
		if lineIdx >= len(buffer.lines) {
			break
		}

		line := buffer.lines[lineIdx]
		line = replaceTabsWithSpaces(line, settings.tabstop, settings.tabchar)

		fmt.Printf("%s", line)
		s.setVisualCursorPosition(s.leftChromeWidth, s.visualCursorY+1)
	}

	// Resetting state after re-draw
	s.setVisualCursorPosition(preDrawCursorX, preDrawCursorY)
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
