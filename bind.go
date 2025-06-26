package bcl

type bindSelector byte

const (
	// must fit in the lower half-byte
	bindOne bindSelector = iota + 1
	bindFirst
	bindLast
	bindNamedBlock
	bindNamedBlocks
	bindAll = 15
)

type bindTarget byte

const (
	// must fit in the upper half-byte
	bindStruct bindTarget = (iota + 1) * 16
	bindSlice
)

type (
	Binding interface{ binding() }

	StructBinding   struct{ Value Block }
	SliceBinding    struct{ Value []Block }
	UmbrellaBinding struct{ Parts []Binding }
)

func (SliceBinding) binding()     {}
func (StructBinding) binding()    {}
func (*UmbrellaBinding) binding() {}
