package main

import (
	"testing"
)

func (p *Program[MockTerminal]) assertLogicalPos(
	t *testing.T,
	x, y int,
) {
	panel := p.state.panels[p.state.activePanelIdx]
	actX, actY := panel.logicalCursorX, panel.logicalCursorY

	if actX != x || actY != y {
		t.Errorf("wanted logical pos x=%d,y=%d; got x=%d,y=%d", x, y, actX, actY)
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

//func TestXTruncatesWhenMoveUpToShorterLine(t *testing.T) {
//	p := testingProgramFromBuf("a\nbc")
//    p.processInputs('j', 'l')
//    p.assertLogicalPos(t, 1, 1)
//    p.processInputs('k')
//    p.assertLogicalPos(t, 0, 0)
//}
