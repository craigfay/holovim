package main

type NormalModeKeyBindings struct {
	cursorUp        rune
	cursorDown      rune
	cursorLeft      rune
	cursorRight     rune
	closeBuffer     rune
	insertLeft      rune
	insertRight     rune
	insertAbove     rune
	insertBelow     rune
	insertLineStart rune
	insertLineEnd   rune
}

var DefaultNormalModeKeyBindings = NormalModeKeyBindings{
	cursorUp:    'k',
	cursorDown:  'j',
	cursorLeft:  'h',
	cursorRight: 'l',
	closeBuffer: 'q',
	insertLeft:  'i',

	// TODO these should probably be macros that move the cursor, then insertLeft
	insertRight:     'a',
	insertAbove:     'o',
	insertBelow:     'O',
	insertLineStart: 'I',
	insertLineEnd:   'A',
}

func normalMode[T Terminal](input rune, prog *Program[T]) {
	keys := &prog.settings.normalModeKeybind

	if input == keys.insertLeft {
		prog.changeMode(InsertMode)
		return
	}

	if input == keys.cursorDown || input == RuneDownArrow {
		prog.moveCursorDown()
	}

	if input == keys.cursorUp || input == RuneUpArrow {
		prog.moveCursorUp()
	}

	if input == keys.cursorLeft || input == RuneLeftArrow {
		prog.moveCursorLeft()
	}

	if input == keys.cursorRight || input == RuneRightArrow {
		prog.moveCursorRight()
	}

	if input == keys.closeBuffer {
		prog.state.shouldExit = true
	}
}
