package main

import (
	"os"
)

type InputIterator interface {
	Next() (bool, byte, error)
}

type StdinIterator struct {
	buf []byte
}

func NewStdinIterator() *StdinIterator {
	return &StdinIterator{
		buf: make([]byte, 1), // Buffer to store a single byte
	}
}

func (it *StdinIterator) Next() (bool, byte, error) {
	_, err := os.Stdin.Read(it.buf)
	if err != nil {
		return false, 0, err
	}
	return false, it.buf[0], nil
}

type StaticInputIterator struct {
	inputs []byte
	index  int
}

func NewStaticInputIterator(inputs []byte) *StaticInputIterator {
	return &StaticInputIterator{
		inputs: inputs,
		index:  0,
	}
}

func (it *StaticInputIterator) Next() (bool, byte, error) {
	if it.index >= len(it.inputs) {
		return true, 0, nil
	}
	input := it.inputs[it.index]
	it.index++
	return false, input, nil
}

