package main

import (
	"unicode"
)

func contains[T comparable](list []T, item T) bool {
	for _, other := range list {
		if item == other {
			return true
		}
	}
	return false
}

func isStandardUnicode(r rune) bool {
	printableControl := []rune{RuneTab}
	return r < 0xE000 && (unicode.IsPrint(r) || contains(printableControl, r))
}

// Defining aliases for regular unicode chars that aren't always easily recognizable
const (
	RuneEscape         rune = '\x1b'
	RuneEnter          rune = '\n'
	RuneCarriageReturn rune = '\r'
	RuneBackspace      rune = '\b'
	RuneDelete         rune = '\x7f'
	RuneTab            rune = '\t'
)

// Mapping common special keys (like arrows, function keys,
// navigation keys) into the Unicode Private Use Area (PUA).
// This ensures they don't collide with printable Unicode chars.
const (
	// Arrow keys
	RuneUpArrow    rune = 0xE000
	RuneDownArrow  rune = 0xE001
	RuneRightArrow rune = 0xE002
	RuneLeftArrow  rune = 0xE003

	// Navigation
	RuneHome     rune = 0xE010
	RuneEnd      rune = 0xE011
	RunePageUp   rune = 0xE012
	RunePageDown rune = 0xE013
	RuneInsert   rune = 0xE014

	// Modifier keys (used as standalone keys)
	RuneCtrl  rune = 0xE020
	RuneAlt   rune = 0xE021
	RuneShift rune = 0xE022

	// Function keys (F1–F12)
	RuneF1  rune = 0xE100
	RuneF2  rune = 0xE101
	RuneF3  rune = 0xE102
	RuneF4  rune = 0xE103
	RuneF5  rune = 0xE104
	RuneF6  rune = 0xE105
	RuneF7  rune = 0xE106
	RuneF8  rune = 0xE107
	RuneF9  rune = 0xE108
	RuneF10 rune = 0xE109
	RuneF11 rune = 0xE10A
	RuneF12 rune = 0xE10B

	// Mouse events
	RuneMouseLeft     rune = 0xE200
	RuneMouseMiddle   rune = 0xE201
	RuneMouseRight    rune = 0xE202
	RuneMouseScrollUp rune = 0xE203
	RuneMouseScrollDn rune = 0xE204
)
