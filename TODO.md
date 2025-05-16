
## Bugs
- Bad error message when incorrect args passed on startup
- End of Input yields an error message that pollutes test output

## Testing Scenarios
- Track drawn content in MockTerminal, and assert against
- Up and down motions preserve visualCursorY, even when logicalCursorY changes.
- Wrapping backwards to the last line doesn't land the cursor on columns that don't exist


## Features
- Preserve X position when moving between lines of different lengths
- Add a setting for whether horizontal cursor movement can overflow and underflow to different lines
- Use a "panel" data model, where each panel tracks its own cursor position
