package bcl

type bindSelector byte

const (
	// must fit in the lower half-byte
	bindOne bindSelector = iota + 1
	bindFirst
	bindLast
	bindNamedBlock
	bindAll = 15
)

type bindTarget byte

const (
	// must fit in the upper half-byte
	bindStruct bindTarget = (iota + 1) * 16
	bindSlice
)

type Binding interface {
	binding()
}

type StructBinding struct{ Value Block }

func (StructBinding) binding() {}

type SliceBinding struct{ Value []Block }

func (SliceBinding) binding() {}
