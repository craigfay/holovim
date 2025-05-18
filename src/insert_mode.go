
package main

func insertMode[T Terminal](input byte, prog *Program[T]) {
	if input == Escape {
		prog.state.currentMode = NormalMode
	}
}
