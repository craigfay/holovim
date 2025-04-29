
## Testing Scenarios
- Can scroll by moving rightward at the end of the last line
- Cannot exceed first or last characters in file with motions
- Up and down motions preserve visualCursorY, even when logicalCursorY changes.
- Wrapping backwards to the last line doesn't land the cursor on columns that don't exist


## Features
- Add a setting for whether horizontal cursor movement can overflow and underflow to different lines
