package main

import (
	"runtime"
	"testing"
)

func failWithStackTrace(t *testing.T, format string, args ...interface{}) {
	stackBuf := make([]byte, 1024)
	stackSize := runtime.Stack(stackBuf, false)
	stackTrace := string(stackBuf[:stackSize])

	format += "\nStack trace:\n%s"
	args = append(args, stackTrace)

	t.Errorf(format, args...)
}

func (p *Program[MockTerminal]) assertLogicalPos(
	t *testing.T,
	x, y int,
) {
	tab := p.state.tabs[p.state.activeTabIdx]
	panel := tab.panels[tab.activePanelIdx]
	actX, actY := panel.logicalCursorX, panel.logicalCursorY

	if actX != x || actY != y {
		failWithStackTrace(t, "wanted logical pos x=%d,y=%d; got x=%d,y=%d", x, y, actX, actY)
	}
}

func (p *Program[MockTerminal]) processInputs(i ...byte) {
	it := NewStaticInputIterator(i)
	runMainLoop(p, it)
}

func TestCursorDown(t *testing.T) {
	p := testingProgramFromBuf(basicBuf)
	p.processInputs('j')
	p.assertLogicalPos(t, 0, 1)
}

func TestCursorUp(t *testing.T) {
	p := testingProgramFromBuf(basicBuf)
	p.processInputs('j', 'k')
	p.assertLogicalPos(t, 0, 0)
}

func TestCursorRight(t *testing.T) {
	p := testingProgramFromBuf(basicBuf)
	p.processInputs('l')
	p.assertLogicalPos(t, 1, 0)
}

func TestCursorLeft(t *testing.T) {
	p := testingProgramFromBuf(basicBuf)
	p.processInputs('l', 'h')
	p.assertLogicalPos(t, 0, 0)
}

func TestRightwardWrapToNextLine(t *testing.T) {
	p := testingProgramFromBuf("ab\ncd")
	p.processInputs('l', 'l')
	p.assertLogicalPos(t, 0, 1)
}

func TestLeftwardWrapToPrevLine(t *testing.T) {
	p := testingProgramFromBuf("ab\ncd")
	p.processInputs('j', 'h')
	p.assertLogicalPos(t, 1, 0)
}

func TestCannotMoveCursorBeforeFirstChar(t *testing.T) {
	p := testingProgramFromBuf("a\nb")
	p.processInputs('h')
	p.assertLogicalPos(t, 0, 0)
	p.processInputs('k')
	p.assertLogicalPos(t, 0, 0)
}

func TestCannotMoveCursorBeyondLastChar(t *testing.T) {
	p := testingProgramFromBuf("a\nb")
	p.processInputs('l', 'l', 'l')
	p.assertLogicalPos(t, 0, 1)
	p.processInputs('j')
	p.assertLogicalPos(t, 0, 1)
}

func TestXTruncatesWhenMoveDownToShorterLine(t *testing.T) {
	p := testingProgramFromBuf("ab\nc")
	p.processInputs('l')
	p.assertLogicalPos(t, 1, 0)
	p.processInputs('j')
	p.assertLogicalPos(t, 0, 1)
}

func TestXTruncatesWhenMoveUpToShorterLine(t *testing.T) {
	p := testingProgramFromBuf("a\nbc")
	p.processInputs('j', 'l')
	p.assertLogicalPos(t, 1, 1)
	p.processInputs('k')
	p.assertLogicalPos(t, 0, 0)
}

func TestPinnedVisualXRespectedByCursorDown(t *testing.T) {
	p := testingProgramFromBuf("aaa\n" + "bb\n" + "cccc")
	p.processInputs('l', 'l') // setup - moving to the end of 1st line
	p.assertLogicalPos(t, 2, 0)
	p.processInputs('j') // moving down to 2nd line
	p.assertLogicalPos(t, 1, 1)
	p.processInputs('j')        // moving down to 3rd line
	p.assertLogicalPos(t, 2, 2) // expecting pinned visual x to be restored
}

func TestPinnedVisualXRespectedByCursorUp(t *testing.T) {
	p := testingProgramFromBuf("aaa\n" + "bb\n" + "cccc")
	p.processInputs('j', 'j', 'l', 'l') // setup - moving to the 3rd column of the last line
	p.assertLogicalPos(t, 2, 2)
	p.processInputs('k', 'k')   // moving back to the first line
	p.assertLogicalPos(t, 2, 0) // expecting pinned visual x to be restored
}
