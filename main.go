package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"strings"
)

type Buffer struct {
	filepath          string
	lines             []string
	topVisibleLineIdx int
}

type Motions struct {
	cursor_up    byte
	cursor_down  byte
	cursor_left  byte
	cursor_right byte
}

type ProgramState struct {
	buffers            []Buffer
	motions            Motions
	activeBufferIdx    int
	needsRedraw        bool
	termHeight         int
	topChromeContent   []string
	topChromeHeight    int
	leftChromeWidth    int
	bottomChromeHeight int
	visualCursorY      int
	visualCursorX      int
	lastVisualCursorY  int
	lastVisualCursorX  int
	logicalCursorX     int
	logicalCursorY     int
	lastLogicalCursorX int
	lastLogicalCursorY int
	// Represents the last visual cursor x that the user
	// has selected. When they move up and down to a line
	// that is shorter than the last one, the visual cursor
	// will change, but we want to restore it whenever moving
	// to a line that does have enough characters.
	bookmarkedVisualCursorX int
}

type Settings struct {
	tabstop           int
	tabchar           string
	cursor_x_overflow bool
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

func main() {
	args := os.Args

	_ = getLogger("./logfile.log.txt")

	// In development, a workaround is necessary to pass arguments
	// that end in ".go", because the go compiler thinks they are part
	// of invalid input, instead of an argument to our compiled program.
	// In this case, we can use `go run main.go -- editme.go`. This
	// codeblock modifies the args list to allow this workaround.
	for i, arg := range args {
		if arg == "--" {
			// Ignoring everything before "--"
			args = args[i:]
			break
		}
	}

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

	s := ProgramState{}

	settings := Settings{
		tabstop:           4,
		tabchar:           "›", // (U+203A)
		cursor_x_overflow: true,
	}

	s.motions = Motions{
		cursor_up:    'k',
		cursor_down:  'j',
		cursor_left:  'h',
		cursor_right: 'l',
	}

	ANSI := ANSIInstructions{}

	s.buffers = []Buffer{
		{
			filepath:          filename,
			lines:             lines,
			topVisibleLineIdx: 0,
		},
	}

	s.activeBufferIdx = 0

	// Saving the current state of the terminal,
	// and re-loading it when this program exits
	oldTerminalState, err := term.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		panic(err)
	}

	// Resetting cursor position and terminal state after the program closes.
	// This isn't exactly true, because we're not resetting it exactly as it was.
	defer ANSI.setCursorPosition(0, 0)
	defer term.Restore(int(os.Stdin.Fd()), oldTerminalState)

	termHeight, _, err := ANSI.getTerminalSize()
	s.termHeight = termHeight

	if err != nil {
		panic(err)
	}

	s.topChromeHeight = 1
	s.bottomChromeHeight = 1

	// 3 columns for line numbers, 2 columns for padding
	s.leftChromeWidth = 5

	contentAreaRowCount := s.termHeight - s.topChromeHeight - s.bottomChromeHeight

	// The maximum allowed cursor_y position inside of the content area
	contentAreaMaxY := s.termHeight - s.bottomChromeHeight

	// The minimum allowed cursor_x position inside of the content area

	s.topChromeContent = []string{
		"Press \"q\" to exit...",
	}

	// Logical cursor position
	s.logicalCursorX = 0
	s.logicalCursorY = 0

	s.lastLogicalCursorX = 0
	s.lastLogicalCursorY = 0

	// Visual cursor position
	s.visualCursorX = s.leftChromeWidth
	s.visualCursorY = s.topChromeHeight

	s.lastVisualCursorX = s.visualCursorX
	s.lastVisualCursorY = s.visualCursorY

	ANSI.setCursorPosition(s.visualCursorX, s.visualCursorY)

	// Adding a helper to deliver ANSI instruction, while
	// also updating native variables to track the cursor
	setVisualCursorPosition := func(x, y int) {
		ANSI.setCursorPosition(x, y)
		s.lastVisualCursorX = s.visualCursorX
		s.lastVisualCursorY = s.visualCursorY
		s.visualCursorX = x
		s.visualCursorY = y
		s.needsRedraw = true
	}

	setLogicalCursorPosition := func(x, y int) {
		s.lastLogicalCursorX = s.logicalCursorX
		s.lastLogicalCursorY = s.logicalCursorY
		s.logicalCursorX = x
		s.logicalCursorY = y
		s.needsRedraw = true
	}

	// Declaring a buffer to store a single byte of user input at a time
	buf := make([]byte, 1)

	s.needsRedraw = true

	ANSI.clearScreen()

	for {
		buffer := &s.buffers[s.activeBufferIdx]
		lineNumber := s.logicalCursorY
		columnNumber := s.logicalCursorX
		lineContent := &buffer.lines[lineNumber]
		lineLength := len(*lineContent)
		isAtEndOfLine := columnNumber+1 >= lineLength
		isLastLine := lineNumber == len(buffer.lines)-1
		isAtViewportBottom := s.visualCursorY == contentAreaMaxY
		isAtContentBottom := s.logicalCursorY+1 >= len(buffer.lines)

		if true || s.needsRedraw {
			preDrawCursorX := s.visualCursorX
			preDrawCursorY := s.visualCursorY

			ANSI.clearScreen()
			setVisualCursorPosition(0, 0)

			// Printing top chrome content
			for i := 0; i < s.topChromeHeight; i++ {
				line := s.topChromeContent[i]

				fmt.Printf("%s", line)
				setVisualCursorPosition(s.leftChromeWidth, s.visualCursorY+1)
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
				setVisualCursorPosition(s.leftChromeWidth, s.visualCursorY+1)
			}

			// Resetting state after re-draw
			setVisualCursorPosition(preDrawCursorX, preDrawCursorY)
			s.needsRedraw = false
		}

		// Reading a single byte from stdin into the buffer
		_, err := os.Stdin.Read(buf)

		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		if buf[0] == s.motions.cursor_down {
			canScroll := buffer.topVisibleLineIdx+contentAreaRowCount+1 < len(buffer.lines)

			if !isAtContentBottom || canScroll {
				// moving the cursor down
				nextLine := buffer.lines[s.logicalCursorY+1]
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

				setVisualCursorPosition(newVisualX, newVisualY)
				setLogicalCursorPosition(newLogicalX, s.logicalCursorY+1)
			}
		}

		if buf[0] == s.motions.cursor_up {
			canScroll := buffer.topVisibleLineIdx > 0

			if s.logicalCursorY > 0 || canScroll {
				prevLine := buffer.lines[s.logicalCursorY-1]
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

				setVisualCursorPosition(newVisualX, newVisualY)
				setLogicalCursorPosition(newLogicalX, s.logicalCursorY-1)
			}
		}

		if buf[0] == s.motions.cursor_left {
			// Wrapping to the end of the previous line
			if columnNumber == 0 && lineNumber != 0 {
				if !settings.cursor_x_overflow {
					continue
				}
				prevLine := buffer.lines[lineNumber-1]
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

				setLogicalCursorPosition(newLogicalX, lineNumber-1)
				setVisualCursorPosition(newVisualX, newVisualY)
				s.bookmarkedVisualCursorX = newVisualX

			} else if columnNumber != 0 {
				// Moving the cursor left within the current line
				thisChar := (*lineContent)[columnNumber-1]
				newVisualX := s.visualCursorX

				if thisChar == '\t' {
					newVisualX -= settings.tabstop
				} else {
					newVisualX -= 1
				}

				setLogicalCursorPosition(columnNumber-1, lineNumber)
				setVisualCursorPosition(newVisualX, s.visualCursorY)
				s.bookmarkedVisualCursorX = newVisualX
			}
		}

		if buf[0] == s.motions.cursor_right {
			if isAtEndOfLine && isLastLine {
				continue
			}

			// wrapping to the beginning of the next line
			if isAtEndOfLine && !isLastLine {
				if !settings.cursor_x_overflow {
					continue
				}

				newVisualY := s.visualCursorY + 1

				// scrolling if necessary
				if isAtViewportBottom {
					newVisualY = s.visualCursorY
					buffer.topVisibleLineIdx += 1
				}

				setLogicalCursorPosition(0, s.logicalCursorY+1)
				setVisualCursorPosition(s.leftChromeWidth, newVisualY)
				s.bookmarkedVisualCursorX = s.leftChromeWidth

				// moving the cursor right
			} else {
				thisChar := (*lineContent)[s.logicalCursorX]
				newVisualX := s.visualCursorX

				if thisChar == '\t' {
					newVisualX += settings.tabstop
				} else {
					newVisualX += 1
				}

				setLogicalCursorPosition(s.logicalCursorX+1, s.logicalCursorY)
				setVisualCursorPosition(newVisualX, s.visualCursorY)
				s.bookmarkedVisualCursorX = newVisualX
			}
		}

		// Exiting the loop if 'q' is pressed
		if buf[0] == 'q' {
			ANSI.clearScreen()
			break
		}

		s.needsRedraw = true
	}
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
