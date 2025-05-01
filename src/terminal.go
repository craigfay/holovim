package main

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

type MockTerminal struct {
	cursorX int
	cursorY int
}

func (t *MockTerminal) printf(s string, args ...interface{}) {}

func (t *MockTerminal) clearScreen() {}

func (t *MockTerminal) setCursorPosition(x, y int) {
	t.cursorX = x
	t.cursorY = y
}

func (t *MockTerminal) getCursorPosition() (x, y int, err error) {
	return t.cursorX, t.cursorY, nil
}

func (t *MockTerminal) getSize() (rows, cols int, err error) {
	return 100, 100, nil
}

type ANSI struct{}

func (ANSI) printf(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}

func (ANSI) clearScreen() {
	fmt.Printf("\x1b[2J")
}

func (ANSI) setCursorPosition(x, y int) {
	// Incrementing the given values, because ANSI row/col positions
	// seem to be 1-indexed instead of 0-indexed
	fmt.Printf("\033[%d;%dH", y+1, x+1)
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
}

func (ANSI) getSize() (rows, cols int, err error) {
	last_x, last_y, err := ANSI{}.getCursorPosition()

	if err != nil {
		return 0, 0, err
	}

	// Moving cursor to bottom-right
	ANSI{}.setCursorPosition(9999, 9999)

	w, h, err := ANSI{}.getCursorPosition()

	if err != nil {
		return 0, 0, err
	}

	ANSI{}.setCursorPosition(last_x, last_y)

	return w, h, nil
}

