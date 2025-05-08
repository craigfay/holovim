package main

import (
    "testing"
)


func (p *Program[MockTerminal]) assertLogicalPos(
    t *testing.T,
    x, y int,
) {
    actX, actY := p.state.logicalCursorX, p.state.logicalCursorY

	if actX != x || actY != y {
		t.Errorf("wanted logical pos x=%d,y=%d; got x=%d,y=%d", x, y, actX, actY)
	}
}

func (p *Program[MockTerminal]) assertVisualPos(
    t *testing.T,
    x, y int,
) {
    actX, actY := p.state.visualCursorX, p.state.visualCursorY

	if actX != x || actY != y {
		t.Errorf("wanted visual pos x=%d,y=%d; got x=%d,y=%d", x, y, actX, actY)
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

