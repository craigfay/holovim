package main

import (
	"runtime"
	"strings"
	"testing"
)

func (prog *Program[MockTerminal]) assertBufferContent(t *testing.T, expected ...string) {
	actual := prog.getActiveBuffer().lines

	for i, expectedLine := range expected {
		actualLine := actual[i]

		if actualLine != expectedLine {
			failWithStackTrace(t, "Line %d:\nWanted: `%v`\nGot: `%v`", i, expectedLine, actualLine)
		}
	}
}

func failWithStackTrace(t *testing.T, format string, args ...interface{}) {
	stackBuf := make([]byte, 1024)
	stackSize := runtime.Stack(stackBuf, false)
	stackTrace := string(stackBuf[:stackSize])

	format += "\nStack trace:\n%s"
	args = append(args, stackTrace)

	t.Errorf(format, args...)
}

func (p *Program[MockTerminal]) processInputs(i ...byte) {
	it := NewStaticInputIterator(i)
	runMainLoop(p, it)
}

func (p *Program[MockTerminal]) assertLogicalPos(
	t *testing.T,
	x, y int,
) {
	tab := p.state.tabs[p.state.activeTabIdx]
	panel := tab.panels[tab.activePanelIdx]
	actX, actY := panel.logicalCursorX, panel.logicalCursorY

	if actX != x || actY != y {
		failWithStackTrace(t, "wanted logical pos x=%d,y=%d; got x=%d,y=%d", x, y, actX, actY)
	}
}

func testingProgramFromBuf(buf string) Program[MockTerminal] {
	buffers := []Buffer{
		{
			filepath:          "test",
			lines:             strings.Split(buf, "\n"),
			topVisibleLineIdx: 0,
		},
	}

	program := Program[MockTerminal]{
		logger:   getLogger("./logfile_test.log.txt"),
		state:    ProgramState{},
		term:     MockTerminal{},
		settings: defaultSettings(),
	}

	program.state.buffers = buffers
	initializeState(&program)
	return program
}

const basicBuf = `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Terminal interface {
	clearScreen()
	setCursorPosition(x, y int)
	getCursorPosition() (x, y int, err error)
	getSize() (rows, cols int, err error)
	printf(s string, args ...interface{})
}

func (ANSI) getCursorPosition() (x, y int, err error) {
	// Querying the terminal for cursor position
	fmt.Print("\033[6n")

	// Reading the response
	var response []byte
	buf := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read from stdin: %v", err)
		}
		if buf[0] == 'R' {
			break
		}
		response = append(response, buf[0])
	}

	// Parsing the response
	// Response format: "\033[<rows>;<cols>R"
	parts := strings.Split(strings.Trim(string(response), "\033["), ";")

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected response format: %s", response)
	}

	rows, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse rows: %v", err)
	}

	cols, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse cols: %v", err)
	}

	return rows, cols, nil
}`
