package main

import (
    "fmt"
    "os"
	"bufio"
	"strconv"
	"strings"
    "golang.org/x/term"
)

// A type that maps human-readable names to their corresponding ANSI escape sequence.
// These can be given to the terminal as strings, which represent special instructions.
// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
type AnsiEscapeSequences struct {
	ClearScreen string
	CursorToPosition func(row, col int) string
	CursorMoveUp func(n int) string
	CursorMoveDown func(n int) string
	CursorMoveRight func(n int) string
	CursorMoveLeft func(n int) string
	CursorToNextLine func(n int) string
	CursorToPrevLine func(n int) string
	CursorToColumn func(n int) string
	HighlightOn string
	HighlightOff string

}

var ansi = AnsiEscapeSequences {
	ClearScreen: "\x1b[2J",
	CursorToPosition: func(row, col int) string { return fmt.Sprintf("\x1b[%d;%dH", row, col) },
	CursorMoveUp: func(n int) string { return fmt.Sprintf("\x1b[%dA", n) },
	CursorMoveDown: func(n int) string { return fmt.Sprintf("\x1b[%dB", n) },
	CursorMoveRight: func(n int) string { return fmt.Sprintf("\x1b[%dC", n) },
	CursorMoveLeft: func(n int) string { return fmt.Sprintf("\x1b[%dD", n) },
	CursorToNextLine: func(n int) string { return fmt.Sprintf("\x1b[%dE", n) },
	CursorToPrevLine: func(n int) string { return fmt.Sprintf("\x1b[%dF", n) },
	CursorToColumn: func(n int) string { return fmt.Sprintf("\x1b[%dG", n) },
	HighlightOn: "\x1b[7m",
	HighlightOff: "\x1b[0m",
}

func clearScreen() {
	fmt.Printf("%s", ansi.ClearScreen)
}

func setCursorPosition(x, y int) {
	// Incrementing the given values, because ANSI row/col positions
	// seem to be 1-indexed instead of 0-indexed
	fmt.Printf("\033[%d;%dH", y + 1, x + 1)
}

func main() {
    // Checking if a filename is provided as a command-line argument
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <filename>")
        return
    }

    // Extracting the filename from the command-line arguments, and opening it
    filename := os.Args[1]
    file, err := os.Open(filename)

    if err != nil {
        fmt.Printf("Error opening file %s: %v\n", filename, err)
        return
    }

    // Ensuring the file is closed when the program exits
    defer file.Close()

	// Loading the file contents into a list of strings, line by line
	lines := []string {}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
        fmt.Printf("Error scanning file %s\n", err)
		return
	}

	// Saving the current state of the terminal,
	// and re-loading it when this program exits
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		panic(err)
	}

	// Resetting cursor position and terminal state after the program closes.
	// This isn't exactly true, because we're not resetting it exactly as it was.
	defer setCursorPosition(0, 0)
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	term_height, term_width, err := getTerminalSize()

	if err != nil {
		panic(err)
	}

	top_chrome_height := 1
	bottom_chrome_height := 1

	// 3 columns for line numbers, 2 columns for padding
	left_chrome_width := 5

	content_area_row_count := term_height - top_chrome_height - bottom_chrome_height

	// The minimum allowed cursor_y position inside of the content area
	content_area_min_y := top_chrome_height

	// The maximum allowed cursor_y position inside of the content area
	content_area_max_y := term_height - bottom_chrome_height

	// The minimum allowed cursor_x position inside of the content area
	content_area_min_x := left_chrome_width

	// The maximum allowed cursor_x position inside of the content area
	content_area_max_x := term_width

	clearScreen()
	setCursorPosition(0, 0)

	// Setting values to track where we believe that the cursor is
	cursor_x := left_chrome_width
	cursor_y := top_chrome_height

	top_chrome_content := []string{
		"Press \"q\" to exit...",
	}

	// Printing top chrome content
	for i := 0; i < top_chrome_height; i++ {
		line := top_chrome_content[i]

		fmt.Printf("%s", line)
		setCursorPosition(left_chrome_width, cursor_y + 1)
		cursor_x = left_chrome_width
		cursor_y += 1
	}

	// Printing main buffer content
	for i := 0; i < content_area_row_count; i++ {
		if i >= len(lines) {
			break
		}

		line := lines[i]

		fmt.Printf("%s", line)
		setCursorPosition(left_chrome_width, cursor_y + 1)
		cursor_x = left_chrome_width
		cursor_y += 1
	}


	setCursorPosition(left_chrome_width, top_chrome_height)
	cursor_y = top_chrome_height
	cursor_x = left_chrome_width

	// Declaring a buffer to store a single byte of user input at a time
	buf := make([]byte, 1)

	for {
		// Reading a single byte from stdin into the buffer
		_, err := os.Stdin.Read(buf)

		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		if buf[0] == 'j' && cursor_y + 1 <= content_area_max_y  {
			fmt.Printf("%s", ansi.CursorMoveDown(1))
			setCursorPosition(cursor_x, cursor_y + 1)
			cursor_y += 1
		}

		if buf[0] == 'k' && cursor_y - 1 >= content_area_min_y {
			fmt.Printf("%s", ansi.CursorMoveUp(1))
			setCursorPosition(cursor_x, cursor_y - 1)
			cursor_y -= 1
		}

		if buf[0] == 'h' && cursor_x - 1 >= content_area_min_x {
			fmt.Printf("%s", ansi.CursorMoveLeft(1))
			setCursorPosition(cursor_x - 1, cursor_y)
			cursor_x -= 1
		}

		if buf[0] == 'l' && cursor_x + 1 <= content_area_max_x {
			fmt.Printf("%s", ansi.CursorMoveRight(1))
			setCursorPosition(cursor_x + 1, cursor_y)
			cursor_x += 1
		}

		// Exiting the loop if 'q' is pressed
		if buf[0] == 'q' {
			clearScreen()
			break
		}
	}

}

func getTerminalSize() (rows, cols int, err error) {
	fmt.Print("\033[9999;9999H") // Move cursor to bottom-right
	fmt.Print("\033[6n")         // Query cursor position

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

	rows, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse rows: %v", err)
	}

	cols, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse cols: %v", err)
	}


	return rows, cols, nil
}


