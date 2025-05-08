
## Bugs

- Bad error message when incorrect args passed on startup
- End of Input yields an error message that pollutes test output
- Clearing the screen on program end seems to not happen uniformly

## Testing Scenarios
- Can scroll by moving rightward at the end of the last line
- Cannot exceed first or last characters in file with motions
- Up and down motions preserve visualCursorY, even when logicalCursorY changes.
- Wrapping backwards to the last line doesn't land the cursor on columns that don't exist


## Features
- Add a setting for whether horizontal cursor movement can overflow and underflow to different lines
