package main

import (
    "fmt"
    "os"
	"bufio"
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

	defer term.Restore(int(os.Stdin.Fd()), oldState)


	// Displaying the contents of the file on the screen.
	// Because we're in raw mode, we need to manually move the cursor around.
	for _, line := range lines {
		fmt.Printf(
			"%s%s%s",
			line,
			ansi.CursorToNextLine(1),
			ansi.CursorToColumn(0),
		)

	}

	// Declaring a buffer to store a single byte of user input at a time
	buf := make([]byte, 1)

	for {
 		// Reading a single byte from stdin into the buffer
		_, err := os.Stdin.Read(buf)

		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		if buf[0] == 'j' {
			fmt.Printf("%s", ansi.CursorMoveDown(1))
		}

		if buf[0] == 'k' {
			fmt.Printf("%s", ansi.CursorMoveUp(1))
		}

		if buf[0] == 'h' {
			fmt.Printf("%s", ansi.CursorMoveLeft(1))
		}

		if buf[0] == 'l' {
			fmt.Printf("%s", ansi.CursorMoveRight(1))
		}

		// Exiting the loop if 'q' is pressed
		if buf[0] == 'q' {
			fmt.Printf("%s", ansi.ClearScreen)
			break
		}
	}

}

