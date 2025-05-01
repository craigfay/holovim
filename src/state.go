
package main


import (
	"bufio"
	"fmt"
	"os"
)

type Settings struct {
	tabstop           int
	tabchar           string
	cursor_x_overflow bool
}

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
	shouldExit         bool
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

// Adding a helper to deliver ANSI instruction, while
// also updating native variables to track the cursor
func (s *ProgramState) setVisualCursorPosition(x, y int) {
	ANSI{}.setCursorPosition(x, y)
	s.lastVisualCursorX = s.visualCursorX
	s.lastVisualCursorY = s.visualCursorY
	s.visualCursorX = x
	s.visualCursorY = y
	s.needsRedraw = true
}

func (s *ProgramState) setLogicalCursorPosition(x, y int) {
	s.lastLogicalCursorX = s.logicalCursorX
	s.lastLogicalCursorY = s.logicalCursorY
	s.logicalCursorX = x
	s.logicalCursorY = y
	s.needsRedraw = true
}

func initializeState(settings *Settings) (s ProgramState) {
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

	s.motions = Motions{
		cursor_up:    'k',
		cursor_down:  'j',
		cursor_left:  'h',
		cursor_right: 'l',
	}

	s.buffers = []Buffer{
		{
			filepath:          filename,
			lines:             lines,
			topVisibleLineIdx: 0,
		},
	}

	s.activeBufferIdx = 0

	termHeight, _, err := ANSI{}.getSize()
	s.termHeight = termHeight

	if err != nil {
		panic(err)
	}

	s.topChromeHeight = 1
	s.bottomChromeHeight = 1

	// 3 columns for line numbers, 2 columns for padding
	s.leftChromeWidth = 5

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

	ANSI{}.setCursorPosition(s.visualCursorX, s.visualCursorY)
	ANSI{}.clearScreen()

	s.needsRedraw = true

	return s
}


