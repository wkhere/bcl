package bcl

// Block is a dynamic result of running BCL [Interpret].
// It can be put into a static structure via [CopyBlocks].
type Block struct {
	Type, Name string
	Fields     map[string]any
}

// Interpret parses and evaluates the BCL input, creating Blocks.
func Interpret(input []byte, opts ...Option) ([]Block, error) {
	cf := makeConfig(opts)
	inputStr := string(input)

	prog, pstats, err := parse(inputStr, cf)
	if cf.stats {
		printPStats(pstats)
	}
	if err != nil {
		return nil, err
	}

	result, xstats, err := execute(prog, cf)
	if cf.stats {
		printXStats(xstats)
	}
	return result, err

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
// case-insentitive.
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
