
## Resources
- [Escape Sequences](https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797)

## TODO Bugs
- Bad error message when incorrect args passed on startup
- End of Input yields an error message that pollutes test output
- Make the logger available in all contexts, and try to avoid passing it around

## TODO Testing Scenarios
- Track drawn content in MockTerminal, and assert against
- Up and down motions preserve visualCursorY, even when logicalCursorY changes.
- Wrapping backwards to the last line doesn't land the cursor on columns that don't exist

## TODO Features
- Each panel should have its own chrome for line numbers, filename, etc..
- Add a setting for whether horizontal cursor movement can overflow and underflow to different lines
- By default, deleting content doesn't put it in the clipboard.
- Have a clipboard with multiple slots. When pasting, allow users to cycle through their recent cuts/copies.
- Yank to the system clipboard by default, or very easily
- Visual mode
- Periodic autosave
