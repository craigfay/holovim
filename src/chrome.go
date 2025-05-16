package main

import (
	"strings"
)

func (prog *Program[T]) updateTopChrome() {
	tabNames := []string{}

	for _, tab := range prog.state.tabs {
		panel := &tab.panels[tab.activePanelIdx]

		buffer := prog.state.buffers[panel.bufferIdx]
		tabName := buffer.filepath

		if false == prog.settings.tabNamesUseFullFileName {
			parts := strings.Split(buffer.filepath, "/")
			tabName = parts[len(parts)-1]
		}

		tabNames = append(tabNames, tabName)
	}

	prog.state.topChromeContent = []string{
		" " + strings.Join(tabNames[:], " "),
	}
}
