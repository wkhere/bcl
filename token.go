package bcl

import "fmt"

type tokenType int

type token struct {
	typ tokenType
	val string
	err error // either val or err is set
	pos int
}

const (
	tFAIL tokenType = iota
	tEOF            // values <= tEOF are finalizers
	tERR

	tINT
	tFLOAT
	tSTR
	tIDENT

	tVAR
	tDEF
	tEVAL
	tPRINT
	tTRUE
	tFALSE
	tNIL

	tEQ // single equal sign, not to be confused with tEE
	tLCURLY
	tRCURLY
	tLPAREN
	tRPAREN
	//tLBRACKET
	//tRBRACKET

	tOR
	tAND
	tNOT

	tEE // equal-equal
	tBE // bang-equal
	tLT // less-than
	tLE // less-or-equal
	tGT // greater-than
	tGE // greater-or-equal
	tPLUS
	tMINUS
	tSTAR
	tSLASH

	tSEMICOLON

	tMAX // used only in the rules table
)

//go:generate stringer -type=tokenType

func (t token) String() string {
	if t.typ == tERR {
		return fmt.Sprintf("{%s %q %d}", t.typ, t.err, t.pos)
	}
	return fmt.Sprintf("{%s %q %d}", t.typ, t.val, t.pos)
}
