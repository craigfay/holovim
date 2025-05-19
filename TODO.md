
## Bugs
- Bad error message when incorrect args passed on startup
- End of Input yields an error message that pollutes test output

## Testing Scenarios
- Track drawn content in MockTerminal, and assert against
- Up and down motions preserve visualCursorY, even when logicalCursorY changes.
- Wrapping backwards to the last line doesn't land the cursor on columns that don't exist

## Features
- Each panel should have its own chrome for line numbers, filename, etc..
- Add a setting for whether horizontal cursor movement can overflow and underflow to different lines
- By default, deleting content doesn't put it in the clipboard.
- Have a clipboard with multiple slots. When pasting, allow users to cycle through their recent cuts/copies.
