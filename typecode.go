package bcl

type typecode byte

const (
	typeNIL typecode = iota
	typeINT
	typeFLOAT
	typeSTR
	typeBOOL
)

//go:generate stringer -type typecode -trimprefix type
