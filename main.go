package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
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

func main() {
	args := os.Args

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

	// The minimum allowed cursor_y position inside of the content area
	content_area_min_y := top_chrome_height

	// The maximum allowed cursor_y position inside of the content area
	content_area_max_y := term_height - bottom_chrome_height

	// The minimum allowed cursor_x position inside of the content area
	content_area_min_x := left_chrome_width

	top_chrome_content := []string{
		"Press \"q\" to exit...",
	}

	ANSI.setCursorPosition(left_chrome_width, top_chrome_height)

	// Setting values to track where we believe that the cursor is.
	cursor_y := top_chrome_height
	cursor_x := left_chrome_width

	// Adding a helper to deliver ANSI instruction, while
	// also updating native variables to track the cursor
	setCursorPosition := func(x, y int) {
		ANSI.setCursorPosition(x, y)
		cursor_x = x
		cursor_y = y
	}

	// Declaring a buffer to store a single byte of user input at a time
	buf := make([]byte, 1)

	needs_redraw := true

	last_chosen_column_number := 0

	ANSI.clearScreen()

	for {
		buffer := &s.buffers[s.active_buffer_idx]

		if needs_redraw {
			pre_draw_cursor_x := cursor_x
			pre_draw_cursor_y := cursor_y

			ANSI.clearScreen()
			setCursorPosition(0, 0)

			// Printing top chrome content
			for i := 0; i < top_chrome_height; i++ {
				line := top_chrome_content[i]

				fmt.Printf("%s", line)
				setCursorPosition(left_chrome_width, cursor_y+1)
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
				fmt.Printf("%s", line)
				setCursorPosition(left_chrome_width, cursor_y+1)
			}

			// Resetting state after re-draw
			setCursorPosition(pre_draw_cursor_x, pre_draw_cursor_y)
			needs_redraw = false
		}

		line_number := buffer.top_visible_line_idx + cursor_y - top_chrome_height
		column_number := cursor_x - left_chrome_width
		line_length := len(buffer.lines[line_number])
		is_at_end_of_line := column_number+1 >= line_length
		is_last_line := line_number == len(buffer.lines)-1

		highest_column_number_on_line := max(line_length-1, 0)

		// Preventing the cursor from being to the right of the end of the current line
		if column_number > line_length {
			last_chosen_column_number = max(column_number, last_chosen_column_number)
			setCursorPosition(left_chrome_width+highest_column_number_on_line, cursor_y)
			continue
		}

		// Restoring last chosen column number, if possible
		if last_chosen_column_number != column_number && column_number != highest_column_number_on_line {
			setCursorPosition(left_chrome_width+min(last_chosen_column_number, highest_column_number_on_line), cursor_y)
			continue
		}

		// Interpreting 9999 as a "magic" number, which indicates
		// that the user moved the cursor left and wrapped around
		// to the end of the previous line. The reason for this is
		// that it prevents the cursor movement logic from having
		// to know the length of the previous line.
		if last_chosen_column_number == 9999 {
			last_chosen_column_number = line_length - 1
		}

		// Reading a single byte from stdin into the buffer
		_, err := os.Stdin.Read(buf)

		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		if buf[0] == s.motions.cursor_down {
			is_at_viewport_bottom := cursor_y == content_area_max_y
			is_at_content_bottom := cursor_y-top_chrome_height+1 >= len(buffer.lines)
			can_scroll := buffer.top_visible_line_idx+content_area_row_count+1 < len(buffer.lines)

			if !is_at_viewport_bottom && !is_at_content_bottom {
				setCursorPosition(cursor_x, cursor_y+1)
			} else if can_scroll {
				buffer.top_visible_line_idx += 1
				needs_redraw = true
			}
		}

		if buf[0] == s.motions.cursor_up {
			is_at_viewport_top := cursor_y == content_area_min_y
			can_scroll := buffer.top_visible_line_idx > 0

			if !is_at_viewport_top {
				setCursorPosition(cursor_x, cursor_y-1)
			} else if can_scroll {
				buffer.top_visible_line_idx -= 1
				needs_redraw = true
			}
		}

		if buf[0] == s.motions.cursor_left {
			if column_number == 0 && line_number != 0 {
				// wrapping to the end of previous line
				setCursorPosition(9999, cursor_y-1)
				last_chosen_column_number = 9999
			} else if column_number != 0 {
				// moving the cursor left
				setCursorPosition(cursor_x-1, cursor_y)
				last_chosen_column_number = cursor_x - left_chrome_width
			}
		}

		if buf[0] == s.motions.cursor_right {
			if is_at_end_of_line && !is_last_line {
				// wrapping to the beginning of the next line
				setCursorPosition(content_area_min_x, cursor_y+1)
				last_chosen_column_number = 0
			} else {
				// moving the cursor right
				setCursorPosition(cursor_x+1, cursor_y)
				last_chosen_column_number = cursor_x - left_chrome_width
			}
		}

		// Exiting the loop if 'q' is pressed
		if buf[0] == 'q' {
			ANSI.clearScreen()
			break
		}
	}
}
