package main

type Program[T Terminal] struct {
	settings Settings
	state    ProgramState
	logger   func(string) error
	term     T
}

type Settings struct {
	tabstop                 int
	tabchar                 string
	cursor_x_overflow       bool
	tabNamesUseFullFileName bool
	normalModeKeybind       NormalModeKeyBindings
}

func defaultSettings() Settings {
	return Settings{
		tabstop:                 4,
		tabchar:                 "›",
		cursor_x_overflow:       true,
		tabNamesUseFullFileName: false,
		normalModeKeybind:       DefaultNormalModeKeyBindings,
	}
}

func (b *Buffer) removeLine(lineNum int) {
	b.lines = append(b.lines[:lineNum], b.lines[lineNum+1:]...)
}

func (b *Buffer) updateLine(lineNum int, content string) {
	b.lines[lineNum] = content
}

func (b *Buffer) insertLine(lineNum int, content string) {
	// TODO handle cases where lineNum is out of bounds
	b.lines = append(b.lines, "")
	copy(b.lines[lineNum+1:], b.lines[lineNum:])
	b.lines[lineNum] = content
}

type Buffer struct {
	filepath          string
	lines             []string
	topVisibleLineIdx int
}

type Tab struct {
	panels         []Panel
	activePanelIdx int
}

type ProgramMode int

const (
	NormalMode ProgramMode = iota
	InsertMode
)

type ProgramState struct {
	shouldExit         bool
	currentMode        ProgramMode
	buffers            []Buffer
	tabs               []Tab
	activeTabIdx       int
	needsRedraw        bool
	termHeight         int
	topChromeContent   []string
	topChromeHeight    int
	leftChromeWidth    int
	bottomChromeHeight int
	visualCursorY      int
	visualCursorX      int
	lastVisualCursorY  int
	lastVisualCursorX  int

	// Represents the last visual cursor x that the user
	// has selected. When they move up and down to a line
	// that is shorter than the last one, the visual cursor
	// will change, but we want to restore it whenever moving
	// to a line that does have enough characters.
	bookmarkedVisualCursorX int
}

type Panel struct {
	topLeftX            int
	topLeftY            int
	logicalCursorX      int
	logicalCursorY      int
	lastLogicalCursorX  int
	lastLogicalCursorY  int
	pinnedVisualCursorX int
	width               int
	height              int
	bufferIdx           int
}

func (prog *Program[T]) changeMode(mode ProgramMode) {
	prog.state.currentMode = mode

	if mode == NormalMode {
		prog.term.useBlockCursor()
	} else if mode == InsertMode {
		prog.term.useBarCursor()
	}
}

// Adding a helper to deliver ANSI instruction, while
// also updating native variables to track the cursor
func (p *Program[T]) setVisualCursorPosition(x, y int) {
	p.term.setCursorPosition(x, y)
	p.state.lastVisualCursorX = p.state.visualCursorX
	p.state.lastVisualCursorY = p.state.visualCursorY
	p.state.visualCursorX = x
	p.state.visualCursorY = y
	p.state.needsRedraw = true
}

// TODO move make this a method of Panel
func (p *Program[T]) setLogicalCursorPosition(x, y int) {
	tab := &p.state.tabs[p.state.activeTabIdx]
	panel := &tab.panels[tab.activePanelIdx]
	panel.lastLogicalCursorX = panel.logicalCursorX
	panel.lastLogicalCursorY = panel.logicalCursorY
	panel.logicalCursorX = x
	panel.logicalCursorY = y
	p.state.needsRedraw = true
}

func initializeState[T Terminal](program *Program[T]) {
	s := &program.state

	s.activeTabIdx = 0

	termHeight, termWidth, err := program.term.getSize()
	s.termHeight = termHeight

	if err != nil {
		panic(err)
	}

	s.topChromeHeight = 1
	s.bottomChromeHeight = 1

	// 3 columns for line numbers, 2 columns for padding
	s.leftChromeWidth = 5

	s.topChromeContent = []string{
		"Press \"q\" to exit...",
	}

	panels := []Panel{
		{
			bufferIdx:          0,
			topLeftX:           s.leftChromeWidth,
			topLeftY:           s.topChromeHeight,
			logicalCursorX:     0,
			logicalCursorY:     0,
			lastLogicalCursorX: 0,
			lastLogicalCursorY: 0,
			width:              termWidth - s.leftChromeWidth,
			height:             termHeight - s.topChromeHeight - s.bottomChromeHeight,
		},
	}

	s.tabs = []Tab{
		{
			panels:         panels,
			activePanelIdx: 0,
		},
	}

	// Visual cursor position
	s.visualCursorX = s.leftChromeWidth
	s.visualCursorY = s.topChromeHeight

	s.lastVisualCursorX = s.visualCursorX
	s.lastVisualCursorY = s.visualCursorY

	program.term.setCursorPosition(s.visualCursorX, s.visualCursorY)
	program.term.clearScreen()

	s.needsRedraw = true
}

func (prog *Program[T]) getActivePanel() *Panel {
	tab := &prog.state.tabs[prog.state.activeTabIdx]
	return &tab.panels[tab.activePanelIdx]
}

func (prog *Program[T]) getActiveBuffer() *Buffer {
	panel := prog.getActivePanel()
	return &prog.state.buffers[panel.bufferIdx]
}
