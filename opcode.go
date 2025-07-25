package bcl

type opcode byte

const (
	opNOP opcode = iota

	opRET
	opPRINT

	opSETLOCAL
	opGETLOCAL

	opDEFBLOCK
	opENDBLOCK
	opSETFIELD
	opGETFIELD

	opCONST

	opNIL
	opZERO
	opONE
	opTRUE
	opFALSE

	opNOT

	opEQ
	opLT
	opGT

	opADD
	opSUB
	opMUL
	opDIV
	opNEG
	opUNPLUS

	opJUMP // jump forward
	opLOOP // jump backward
	opJFALSE
	opPOP
	opPOPN

	// since bytecode 2.0:
	opBIND
	opDEFUBIND
	opENDUBIND
)

//go:generate stringer -type opcode -trimprefix op
