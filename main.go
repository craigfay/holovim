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
	filepath             string
	lines                []string
	top_visible_line_idx int
}

type Motions struct {
	cursor_up    byte
	cursor_down  byte
	cursor_left  byte
	cursor_right byte
}

type ProgramState struct {
	buffers           []Buffer
	motions           Motions
	active_buffer_idx int
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

	logger := getLogger("./logfile.log.txt")

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

	s.motions = Motions{
		cursor_up:    'k',
		cursor_down:  'j',
		cursor_left:  'h',
		cursor_right: 'l',
	}

	ANSI := ANSIInstructions{}

	s.buffers = []Buffer{
		{
			filepath:             filename,
			lines:                lines,
			top_visible_line_idx: 0,
		},
	}

	s.active_buffer_idx = 0

	// Saving the current state of the terminal,
	// and re-loading it when this program exits
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		panic(err)
	}

	// Resetting cursor position and terminal state after the program closes.
	// This isn't exactly true, because we're not resetting it exactly as it was.
	defer ANSI.setCursorPosition(0, 0)
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	term_height, _, err := ANSI.getTerminalSize()

	if err != nil {
		panic(err)
	}

	top_chrome_height := 1
	bottom_chrome_height := 1

	// 3 columns for line numbers, 2 columns for padding
	left_chrome_width := 5

	content_area_row_count := term_height - top_chrome_height - bottom_chrome_height

	//content_area_min_y := top_chrome_height

	// The maximum allowed cursor_y position inside of the content area
	content_area_max_y := term_height - bottom_chrome_height

	// The minimum allowed cursor_x position inside of the content area
	//content_area_min_x := left_chrome_width

	top_chrome_content := []string{
		"Press \"q\" to exit...",
	}

	// Logical cursor position
	logicalCursorX := 0
	logicalCursorY := 0

	lastLogicalCursorX := 0
	lastLogicalCursorY := 0

	// Visual cursor position
	visualCursorX := left_chrome_width
	visualCursorY := top_chrome_height

	lastVisualCursorX := visualCursorX
	lastVisualCursorY := visualCursorY

	if lastVisualCursorX == 0 {

	}
	if lastVisualCursorY == 0 {

	}
	if lastLogicalCursorX == 0 {

	}
	if lastLogicalCursorY == 0 {

	}

	ANSI.setCursorPosition(visualCursorX, visualCursorY)

	// Adding a helper to deliver ANSI instruction, while
	// also updating native variables to track the cursor
	setVisualCursorPosition := func(x, y int) {
		ANSI.setCursorPosition(x, y)
		lastVisualCursorX = visualCursorX
		lastVisualCursorY = visualCursorY
		visualCursorX = x
		visualCursorY = y
	}

	setLogicalCursorPosition := func(x, y int) {
		lastLogicalCursorX = logicalCursorX
		lastLogicalCursorY = logicalCursorY
		logicalCursorX = x
		logicalCursorY = y
	}

	// Declaring a buffer to store a single byte of user input at a time
	buf := make([]byte, 1)

	needs_redraw := true

	last_chosen_column_number := 0

	ANSI.clearScreen()

	tabstop := 4

	for {
		buffer := &s.buffers[s.active_buffer_idx]

		line_number := logicalCursorY
		column_number := logicalCursorX
		line_content := &buffer.lines[line_number]
		line_length := len(*line_content)
		is_at_end_of_line := column_number+1 >= line_length
		is_last_line := line_number == len(buffer.lines)-1
		//highest_column_number_on_line := max(line_length-1, 0)

		//newVisualX := left_chrome_width
		//newVisualY := top_chrome_height + logicalCursorY

		//if logicalCursorX != 0 {
		//	for i := 0; i < logicalCursorX; i++ {
		//		if (*line_content)[i] == '\t' {
		//			newVisualX += tabstop
		//		} else {
		//			newVisualX += 1
		//		}
		//	}
		//}

		//setVisualCursorPosition(newVisualX, newVisualY)

		top_chrome_content = []string{
			fmt.Sprintf(
				"chosen_x: %d, x: %d, y: %d, last_x: %d, last_y: %d, vx: %d, vy: %d, last_vx: %d, last_vy: %d, line_len: %d",
				last_chosen_column_number,
				logicalCursorX,
				logicalCursorY,
				lastVisualCursorX,
				lastVisualCursorY,
				visualCursorX,
				visualCursorY,
				lastVisualCursorX,
				lastVisualCursorY,
				line_length,
			),
		}

		if true || needs_redraw {
			pre_draw_cursor_x := visualCursorX
			pre_draw_cursor_y := visualCursorY

			ANSI.clearScreen()
			setVisualCursorPosition(0, 0)

			// Printing top chrome content
			for i := 0; i < top_chrome_height; i++ {
				line := top_chrome_content[i]

				fmt.Printf("%s", line)
				setVisualCursorPosition(left_chrome_width, visualCursorY+1)
			}

			// Printing main buffer content
			for i := 0; i <= content_area_row_count; i++ {
				line_idx := i + buffer.top_visible_line_idx

				// Stopping if about to try to draw a line
				// that doesn't exist
				if line_idx >= len(buffer.lines) {
					break
				}

				line := buffer.lines[line_idx]
				line = replaceTabsWithSpaces(line, tabstop)

				fmt.Printf("%s", line)
				setVisualCursorPosition(left_chrome_width, visualCursorY+1)
			}

			// Resetting state after re-draw
			setVisualCursorPosition(pre_draw_cursor_x, pre_draw_cursor_y)
			needs_redraw = false
		}

		// Reading a single byte from stdin into the buffer
		_, err := os.Stdin.Read(buf)

		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		if buf[0] == s.motions.cursor_down {
			is_at_viewport_bottom := visualCursorY == content_area_max_y
			is_at_content_bottom := logicalCursorY+1 >= len(buffer.lines)
			can_scroll := buffer.top_visible_line_idx+content_area_row_count+1 < len(buffer.lines)

			if !is_at_content_bottom || can_scroll {
				// moving the cursor down
				nextLine := buffer.lines[logicalCursorY+1]
				newLogicalX := 0
				newVisualX := left_chrome_width

				// Incrementing newLogicalX until another increment would
				// exceed the previous visualCursorX
				for {
					if newLogicalX+1 >= len(nextLine) {
						logger(">= len nextline")
						break
					}

					if newVisualX >= visualCursorX {
						logger("> visualCursorX")
						break
					}

					visualXChunk := 0

					isTab := nextLine[newLogicalX] == '\t'

					if isTab {
						visualXChunk += tabstop
					} else {
						visualXChunk += 1
					}

					if newVisualX+visualXChunk > visualCursorX {
						logger(fmt.Sprintf("chunk overflows: newVisualX: %d, visualXChunk: %d, visualCursorY: %d", newVisualX, visualXChunk, visualCursorX))
						break
					}

					newVisualX += visualXChunk
					newLogicalX += 1

					logger(fmt.Sprintf("newVisualX: %d", newVisualX))
				}

				newVisualY := visualCursorY+1

				// Scrolling if necessary
				if is_at_viewport_bottom {
					buffer.top_visible_line_idx += 1
					newVisualY = visualCursorY
				}

				setVisualCursorPosition(newVisualX, newVisualY)
				setLogicalCursorPosition(newLogicalX, logicalCursorY+1)
			}
		}

		if buf[0] == s.motions.cursor_up {
			can_scroll := buffer.top_visible_line_idx > 0

			if logicalCursorY != 0 {

			} else if can_scroll {

			}
		}

		if buf[0] == s.motions.cursor_left {
			if column_number == 0 && line_number != 0 {
				// wrapping to the end of previous line
			} else if column_number != 0 {
				// moving the cursor left
			}
		}

		if buf[0] == s.motions.cursor_right {
			if is_at_end_of_line && is_last_line {
				continue
			}

			if is_at_end_of_line && !is_last_line {
				// wrapping to the beginning of the next line
			} else {
				// moving the cursor right
				thisChar := (*line_content)[logicalCursorX]
				newVisualX := visualCursorX

				if thisChar == '\t' {
					newVisualX += tabstop
				} else {
					newVisualX += 1
				}

				setLogicalCursorPosition(logicalCursorX+1, logicalCursorY)
				setVisualCursorPosition(newVisualX, visualCursorY)
			}
		}

		// Exiting the loop if 'q' is pressed
		if buf[0] == 'q' {
			ANSI.clearScreen()
			break
		}

		needs_redraw = true
	}
}

func replaceTabsWithSpaces(line string, tabWidth int) string {
	spaces := ""
	for i := 0; i < tabWidth; i++ {
		spaces += " "
	}
	return strings.ReplaceAll(line, "\t", spaces)
}
