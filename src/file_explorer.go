package main

import (
	"fmt"
	"os"
)

func dirContentsAsBufferLines(dirPath string) ([]BufferLine, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Creating a slice of BufferLine structs
	var bufferLines []BufferLine
	for _, entry := range entries {
		// Creating a BufferLine for each entry
		line := BufferLine{
			content: entry.Name(),
		}
		line.flags.isDir = entry.IsDir()

		// Adding the BufferLine to the list
		bufferLines = append(bufferLines, line)
	}

	return bufferLines, nil
}
