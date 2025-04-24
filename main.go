package main

import (
    "bufio"
    "fmt"
    "golang.org/x/term"
    "os"
    "strconv"
    "strings"
)

type Buffer struct {
    filepath string
    lines []string
    top_visible_line_idx uint
    bottom_visible_line_idx uint
}

func clearScreen() {
    fmt.Printf("\x1b[2J")
}

func setCursorPosition(x, y int) {
    // Incrementing the given values, because ANSI row/col positions
    // seem to be 1-indexed instead of 0-indexed
    fmt.Printf("\033[%d;%dH", y+1, x+1)
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

    var buffers []Buffer = []Buffer{}

    buffer := Buffer {
        filepath: filename,
        lines: lines,
        top_visible_line_idx: 0,
        bottom_visible_line_idx: 0,
    }

    buffers = append(buffers, buffer)

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

    top_chrome_content := []string{
        "Press \"q\" to exit...",
    }

    setCursorPosition(left_chrome_width, top_chrome_height)

    // Setting values to track where we believe that the cursor is
    cursor_y := top_chrome_height
    cursor_x := left_chrome_width

    // Declaring a buffer to store a single byte of user input at a time
    buf := make([]byte, 1)

    needs_redraw := true

    clearScreen()

    for {
        if needs_redraw {
            pre_draw_cursor_x := cursor_x
            pre_draw_cursor_y := cursor_y

            clearScreen()
            setCursorPosition(0, 0)

            // Printing top chrome content
            for i := 0; i < top_chrome_height; i++ {
                line := top_chrome_content[i]

                fmt.Printf("%s", line)
                setCursorPosition(left_chrome_width, cursor_y+1)
                cursor_x = left_chrome_width
                cursor_y += 1
            }

            // Printing main buffer content
            for i := 0; i < content_area_row_count; i++ {
                if i >= len(buffer.lines) {
                    break
                }

                line := buffer.lines[i]

                fmt.Printf("%s", line)
                setCursorPosition(left_chrome_width, cursor_y+1)
                cursor_x = left_chrome_width
                cursor_y += 1
            }

            // Resetting state after re-draw
            cursor_x = pre_draw_cursor_x
            cursor_y = pre_draw_cursor_y
            needs_redraw = false
        }


        // Reading a single byte from stdin into the buffer
        _, err := os.Stdin.Read(buf)

        if err != nil {
            fmt.Println("Error reading input:", err)
            break
        }

        if buf[0] == 'j' && cursor_y+1 <= content_area_max_y {
            setCursorPosition(cursor_x, cursor_y+1)
            cursor_y += 1
        }

        if buf[0] == 'k' && cursor_y-1 >= content_area_min_y {
            setCursorPosition(cursor_x, cursor_y-1)
            cursor_y -= 1
        }

        if buf[0] == 'h' && cursor_x-1 >= content_area_min_x {
            setCursorPosition(cursor_x-1, cursor_y)
            cursor_x -= 1
        }

        if buf[0] == 'l' && cursor_x+1 <= content_area_max_x {
            setCursorPosition(cursor_x+1, cursor_y)
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
