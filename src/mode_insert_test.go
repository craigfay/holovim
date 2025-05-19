package main

import (
	"testing"
)

func TestBasicInsertion(t *testing.T) {
	p := testingProgramFromBuf("b")
	p.processInputs('i', 'a')
	p.assertBufferContent(t, "ab")
}

func TestBasicBackspace(t *testing.T) {
	p := testingProgramFromBuf("abc")
	p.processInputs('l', 'i', Backspace)
	p.assertBufferContent(t, "bc")
}

func TestBackspaceDoesNothingAtFirstLineFirstChar(t *testing.T) {
	p := testingProgramFromBuf("abc")
	p.processInputs('i', Backspace)
	p.assertBufferContent(t, "abc")
}

func TestBackspaceCanJoinLines(t *testing.T) {
	p := testingProgramFromBuf("abc\n" + "def")
	p.processInputs('j', 'i', Backspace)
	p.assertBufferContent(t, "abcdef")
}
