package bcl

// Unmarshal interprets the BCL input, and stores the result in dest,
// which should be a slice of structs.
// See [AppendBlocks] for a struct format.
func Unmarshal(input []byte, dest any) error {
	res, err := Interpret(input)
	if err != nil {
		return err
	}
	return AppendBlocks(dest, res)
}

// Block is a dynamic result of running BCL [Interpret].
// It can be put into a static structure via [AppendBlocks].
type Block struct {
	Type, Name string
	Fields     map[string]any
}

// Interpret parses and evaluates the BCL input, creating Blocks.
func Interpret(input []byte) ([]Block, error) {
	top, err := parse(input)
	if err != nil {
		return nil, err

	}
	return eval(&top)
}

// AppendBlocks adds the blocks to the dest, which needs to be a pointer
// to a slice of structs.
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
// case-insentitive.
//
// The lack of corresponding fields in the Go struct is reported as error.
// So is type mismatch of the fields.
func AppendBlocks(dest any, blocks []Block) error {
	return appendBlocks(dest, blocks)
}
