// Package bcl provides interpretation of the Basic Configuration Language
// and storing the evaluated result in dynamic Blocks or static structs.
//
//   - [Interpret] parses then executes definitions from a BCL file, then creates
//     Blocks
//   - [CopyBlocks] takes Blocks and saves the content in static Go structs
//   - [Unmarshal] = [Interpret] + [CopyBlocks]
package bcl
