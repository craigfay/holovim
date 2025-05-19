package main

import (
	"testing"
)

func TestBasicInsertion(t *testing.T) {
	p := testingProgramFromBuf("b")
	p.processInputs('i', 'a')
	p.assertBufferContent(t, "ab")
}

