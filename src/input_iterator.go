package main

import (
	"bufio"
	"io"
	"os"
	"time"
)

type InputIterator interface {
	Next() (bool, rune, error)
}

type StdinIterator struct {
	reader     *bufio.Reader
	runes      chan rune     // Channel to store incoming runes
	done       chan struct{} // Channel to signal the goroutine to stop
	tempBuffer []rune        // Buffer to hold runes during escape sequence detection
}

func NewStdinIterator() *StdinIterator {
	it := &StdinIterator{
		reader:     bufio.NewReader(os.Stdin),
		runes:      make(chan rune, 100), // Buffered channel with capacity 100
		done:       make(chan struct{}),
		tempBuffer: []rune{},
	}

	// Starting a goroutine to read runes
	// and send them into the channel
	go func() {
		for {
			select {
			case <-it.done:
				// Stopping the goroutine when signaled
				close(it.runes) // Closing the runes channel to signal EOF
				return
			default:
				r, _, err := it.reader.ReadRune()
				if err != nil {
					// Closing the runes channel on EOF
					if err == io.EOF {
						close(it.runes)
						return
					}
					continue
				}

				// Sending the rune into the channel
				it.runes <- r
			}
		}
	}()

	return it
}

func (it *StdinIterator) Next() (done bool, r rune, err error) {
	// If there are runes in the tempBuffer, returning them first
	if len(it.tempBuffer) > 0 {
		r = it.tempBuffer[0]
		// Removing the returned rune from the buffer
		it.tempBuffer = it.tempBuffer[1:]
		return false, r, nil
	}

	for {
		// Reading the next rune from the channel
		r, ok := <-it.runes
		if !ok {
			// Channel is closed, no more input
			return true, 0, io.EOF
		}

		// If the rune is not ESC, returning it immediately
		if r != '\x1b' {
			return false, r, nil
		}

		// Handle possible escape sequence
		it.tempBuffer = append(it.tempBuffer, r)     // Add ESC to the buffer
		timer := time.NewTimer(1 * time.Millisecond) // Short timeout for escape sequences

		for {
			select {
			case nextRune, ok := <-it.runes:
				if !ok {
					// Channel is closed, no more input
					return true, 0, io.EOF
				}

				it.tempBuffer = append(it.tempBuffer, nextRune) // Adding the rune to the buffer

				// Checking if this is part of a known escape sequence
				if len(it.tempBuffer) == 2 && it.tempBuffer[1] == '[' {
					// Possible CSI sequence (ESC [)
					continue
				}

				if len(it.tempBuffer) == 3 {
					// Handling specific CSI sequences
					switch it.tempBuffer[2] {
					case 'A':
						it.tempBuffer = nil // Clearing the buffer after handling
						return false, RuneUpArrow, nil
					case 'B':
						it.tempBuffer = nil
						return false, RuneDownArrow, nil
					case 'C':
						it.tempBuffer = nil
						return false, RuneRightArrow, nil
					case 'D':
						it.tempBuffer = nil
						return false, RuneLeftArrow, nil
					case 'H':
						it.tempBuffer = nil
						return false, RuneHome, nil
					case 'F':
						it.tempBuffer = nil
						return false, RuneEnd, nil
					case '2':
						if tilde, ok := <-it.runes; ok && tilde == '~' {
							it.tempBuffer = nil
							return false, RuneInsert, nil
						}
					case '3':
						if tilde, ok := <-it.runes; ok && tilde == '~' {
							it.tempBuffer = nil
							return false, RuneDelete, nil
						}
					case '5':
						if tilde, ok := <-it.runes; ok && tilde == '~' {
							it.tempBuffer = nil
							return false, RunePageUp, nil
						}
					case '6':
						if tilde, ok := <-it.runes; ok && tilde == '~' {
							it.tempBuffer = nil
							return false, RunePageDown, nil
						}
					}
				}

				// If it's not part of a known escape sequence, breaking out
				break

			case <-timer.C:
				// Timer expired, treating ESC as a standalone key
				timer.Stop()
				r = it.tempBuffer[0]
				it.tempBuffer = it.tempBuffer[1:] // Removing the returned rune from the buffer
				return false, r, nil
			}
		}

		// If we exit the inner loop without returning, flushing the buffer
		if len(it.tempBuffer) > 0 {
			r = it.tempBuffer[0]
			it.tempBuffer = it.tempBuffer[1:] // Removing the returned rune from the buffer
			return false, r, nil
		}
	}
}

type StaticInputIterator struct {
	inputs []rune
	index  int
}

func NewStaticInputIterator(inputs []rune) *StaticInputIterator {
	return &StaticInputIterator{
		inputs: inputs,
		index:  0,
	}
}

func (it *StaticInputIterator) Next() (bool, rune, error) {
	if it.index >= len(it.inputs) {
		return true, 0, nil
	}
	input := it.inputs[it.index]
	it.index++
	return false, input, nil
}
