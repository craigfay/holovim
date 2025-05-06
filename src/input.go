package main

import (
	"fmt"
	"os"
)

type InputIterator interface {
	Next() (byte, error)
}

type StdinIterator struct {
	buf []byte
}

func NewStdinIterator() *StdinIterator {
	return &StdinIterator{
		buf: make([]byte, 1), // Buffer to store a single byte
	}
}

func (it *StdinIterator) Next() (byte, error) {
	_, err := os.Stdin.Read(it.buf)
	if err != nil {
		return 0, err
	}
	return it.buf[0], nil
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

func (it *StaticInputIterator) Next() (byte, error) {
	if it.index >= len(it.inputs) {
		return 0, fmt.Errorf("end of input")
	}
	input := it.inputs[it.index]
	it.index++
	return input, nil
}
