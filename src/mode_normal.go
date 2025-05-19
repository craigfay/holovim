package main

type NormalModeKeyBindings struct {
	cursorUp        byte
	cursorDown      byte
	cursorLeft      byte
	cursorRight     byte
	closeBuffer     byte
	insertLeft      byte
	insertRight     byte
	insertAbove     byte
	insertBelow     byte
	insertLineStart byte
	insertLineEnd   byte
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

func normalMode[T Terminal](input byte, prog *Program[T]) {
	keys := &prog.settings.normalModeKeybind

	if input == keys.insertLeft {
		prog.changeMode(InsertMode)
		return
	}

	if input == keys.cursorDown {
		prog.moveCursorDown()
	}

	if input == keys.cursorUp {
		prog.moveCursorUp()
	}

	if input == keys.cursorLeft {
		prog.moveCursorLeft()
	}

	if input == keys.cursorRight {
		prog.moveCursorRight()
	}

	if input == keys.closeBuffer {
		prog.state.shouldExit = true
	}
}
