// Package bcl provides interpreting of the Basic Configuration Language (BCL)
// and storing the evaluated result in dynamic Blocks or static structs.
//
//   - [Interpret] parses and executes definitions from a BCL file,
//     then creates Blocks
//   - [Interpret] or [InterpretFile] parses and executes definitions
//     from a BCL file, then creates Blocks
//   - [CopyBlocks] takes Blocks and saves the content in static Go structs
//   - [Unmarshal] = [Interpret] + [CopyBlocks]
//   - [UnmarshalFile] = [InterpretFile] + [CopyBlocks]
package bcl

import "io"

// Block is a dynamic result of running BCL [Interpret].
// It can be put into a static structure via [CopyBlocks].
type Block struct {
	Type, Name string
	Fields     map[string]any
}

// Interpret parses and executes the BCL input, creating Blocks.
func Interpret(input []byte, opts ...Option) ([]Block, error) {
	return interpret(input, "input", opts...)
}

func interpret(input []byte, name string, opts ...Option) (
	[]Block, error,
) {
	cf := makeConfig(opts)
	inputStr := string(input)

	prog, pstats, err := parse(inputStr, name, cf)

	if cf.stats {
		printPStats(cf.output, pstats)
	}
	if err != nil {
		return nil, err
	}

	result, xstats, err := execute(prog, cf)
	if cf.stats {
		printXStats(cf.output, xstats)
	}
	return result, err

}

// FileInput abstracts the input that is read from a file.
// It is going to be closed as soon as it's read.
// The only information needed from a file besides reading/closing
// is that it has a name.
type FileInput interface {
	io.ReadCloser
	Name() string
}

// InterpretFile reads, parses and evaluates the input from a BCL file.
// The file will be closed as soon as possible.
func InterpretFile(f FileInput, opts ...Option) ([]Block, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()
	return interpret(b, f.Name(), opts...)
}

// CopyBlocks copies the blocks to the dest,
// which needs to be a pointer to a slice of structs.
//
// The requirements for the struct are:
//   - struct type name should correspond to the BCL block type
//   - struct needs the Name string field
//   - for each block field, struct needs a corresponding field, of type as
//     the evaluated value (currently supporting int, string and bool)
//
// The mentioned name correspondence is similar to handling json:
// as BCL is expected to use snake case, and Go struct - capitalized camel case,
// the snake underscores are simply removed and then the strings are compared,
// case-insensitive.
//
// The name corresponcence can be also set explicitly,
// by typing a tag keyed `bcl` at the struct field:
//
//	type Record struct {
//		Status string `bcl:"my_status"`
//	}
//
// The lack of corresponding fields in the Go struct is reported as error.
// So is type mismatch of the fields.
//
// If the slice pointed by dest contained any elements, they are overwritten.
func CopyBlocks(dest any, blocks []Block) error {
	return copyBlocks(dest, blocks)
}

// Unmarshal interprets the BCL input, and stores the result in dest,
// which should be a slice of structs.
// See [CopyBlocks] for a struct format.
func Unmarshal(input []byte, dest any) error {
	res, err := Interpret(input)
	if err != nil {
		return err
	}
	return CopyBlocks(dest, res)
}

// UnmarshalFile interprets the BCL file and stores the result in dest,
// which should be a slice of structs.
// See [CopyBlocks] for a struct format.
func UnmarshalFile(f FileInput, dest any) error {
	res, err := InterpretFile(f)
	if err != nil {
		return err
	}
	return CopyBlocks(dest, res)
}
