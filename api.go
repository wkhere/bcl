package bcl

import "fmt" //tmp

type Block struct {
	Kind, Name string
	Fields     map[string]any
}

func Interp(input []byte) error { //...
	top, err := parse(input)
	if err != nil {
		return err

	}
	fmt.Println("vars:")
	fmt.Printf("\t%v\n", top.vars)
	fmt.Println("blocks:")
	for _, x := range top.blocks {
		fmt.Printf("\t%+v\n", x)
	}
	fmt.Println()
	//...
	return nil
}
