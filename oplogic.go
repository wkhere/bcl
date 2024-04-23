package bcl

func binopNumeric(op opcode, a, b value) value {
	switch va := a.(type) {
	case int:
		switch vb := b.(type) {
		case int:
			switch op {
			case opEQ:
				return va == vb
			case opLT:
				return va < vb
			case opGT:
				return va > vb

			case opADD:
				return va + vb
			case opSUB:
				return va - vb
			case opMUL:
				return va * vb
			case opDIV:
				return va / vb
			}
		case float64:
			ca := float64(va)
			switch op {
			case opEQ:
				return ca == vb
			case opLT:
				return ca < vb
			case opGT:
				return ca > vb

			case opADD:
				return ca + vb
			case opSUB:
				return ca - vb
			case opMUL:
				return ca * vb
			case opDIV:
				return ca / vb
			}

		}
	case float64:
		var cb float64
		switch vb := b.(type) {
		case float64:
			cb = vb
		case int:
			cb = float64(vb)
		}
		switch op {
		case opEQ:
			return va == cb
		case opLT:
			return va < cb
		case opGT:
			return va > cb

		case opADD:
			return va + cb
		case opSUB:
			return va - cb
		case opMUL:
			return va * cb
		case opDIV:
			return va / cb
		}

	}

	return nil
}

func unopNumeric(op opcode, a value) value {
	switch va := a.(type) {
	case int:
		switch op {
		case opNEG:
			return -va
		}
	case float64:
		switch op {
		case opNEG:
			return -va
		}
	}
	return nil
}

func binopString(op opcode, a, b string) value {
	switch op {
	case opLT:
		return a < b
	case opGT:
		return a > b

	case opADD:
		return a + b
	}

	return nil
}
