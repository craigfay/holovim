package main

type Program[T Terminal] struct {
	settings Settings
	state    ProgramState
	logger   func(string) error
	term     T
}

type Settings struct {
	tabstop           int
	tabchar           string
	cursor_x_overflow bool
	keybind           KeyBindings
}

func defaultSettings() Settings {
	return Settings{
		tabstop:           4,
		tabchar:           "›",
		cursor_x_overflow: true,
		keybind: KeyBindings{
			cursor_up:    'k',
			cursor_down:  'j',
			cursor_left:  'h',
			cursor_right: 'l',
			close_buffer: 'q',
		},
	}
}

type Buffer struct {
	filepath          string
	lines             []string
	topVisibleLineIdx int
}

type KeyBindings struct {
	cursor_up    byte
	cursor_down  byte
	cursor_left  byte
	cursor_right byte
	close_buffer byte
}

type ProgramState struct {
	shouldExit         bool
	buffers            []Buffer
	activeBufferIdx    int
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
	logicalCursorX     int
	logicalCursorY     int
	lastLogicalCursorX int
	lastLogicalCursorY int
	// Represents the last visual cursor x that the user
	// has selected. When they move up and down to a line
	// that is shorter than the last one, the visual cursor
	// will change, but we want to restore it whenever moving
	// to a line that does have enough characters.
	bookmarkedVisualCursorX int
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

func (p *Program[T]) setLogicalCursorPosition(x, y int) {
	p.state.lastLogicalCursorX = p.state.logicalCursorX
	p.state.lastLogicalCursorY = p.state.logicalCursorY
	p.state.logicalCursorX = x
	p.state.logicalCursorY = y
	p.state.needsRedraw = true
}

func initializeState[T Terminal](program *Program[T]) {
	s := &program.state

	s.activeBufferIdx = 0

	termHeight, _, err := program.term.getSize()
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

	// Logical cursor position
	s.logicalCursorX = 0
	s.logicalCursorY = 0

	s.lastLogicalCursorX = 0
	s.lastLogicalCursorY = 0

	// Visual cursor position
	s.visualCursorX = s.leftChromeWidth
	s.visualCursorY = s.topChromeHeight

	s.lastVisualCursorX = s.visualCursorX
	s.lastVisualCursorY = s.visualCursorY

	program.term.setCursorPosition(s.visualCursorX, s.visualCursorY)
	program.term.clearScreen()

	s.needsRedraw = true
}

