package main

import (
	"fmt"
	"strings"
)

type parsedArgs struct {
	file string

	disasm bool
	trace  bool
	result bool
	stats  bool
	bdump  bool
	bload  bool

	bdumpFile string
	bloadFile string

	help func()
}

const usage = "usage: bcl" +
	" [-d|--disasm] [-t|--trace] [-r|--result] [-s|--stats]" +
	" [--bdump|--bdump=BFILE] [--bload|--bload=BFILE]" +
	" [FILE|-]"

func parseArgs(args []string) (a parsedArgs, _ error) {
	var rest []string
flags:
	for ; len(args) > 0; args = args[1:] {
		switch arg := args[0]; {

		case arg == "-h":
			a.help = func() { fmt.Println(usage) }
			return a, nil

		case arg == "-d", arg == "--disasm":
			a.disasm = true
			continue

		case arg == "-t", arg == "--trace":
			a.trace = true
			continue

		case arg == "-r", arg == "--result":
			a.result = true
			continue

		case arg == "-s", arg == "--stats":
			a.stats = true
			continue

		case strings.HasPrefix(arg, "--bdump"):
			a.bdump = true
			s := arg[len("--bdump"):]
			if len(s) > 0 {
				if s[0] != '=' {
					return a, fmt.Errorf("unknown flag: %s\n%s", arg, usage)
				}
				a.bdumpFile = s[1:]
			}
			continue

		case strings.HasPrefix(arg, "--bload"):
			a.bload = true
			s := arg[len("--bload"):]
			if len(s) > 0 {
				if s[0] != '=' {
					return a, fmt.Errorf("unknown flag: %s\n%s", arg, usage)
				}
				a.bloadFile = s[1:]
			}
			continue

		case arg == "--":
			rest = append(rest, args[1:]...)
			break flags

		case len(arg) > 2 && arg[0] == '-':
			var nonLetter bool
			var oneLetterFlags []string
			for _, c := range arg[1:] {
				if c >= 'a' && c <= 'z' {
					oneLetterFlags = append(oneLetterFlags, "-"+string(c))
				} else {
					nonLetter = true
					break
				}
			}
			if nonLetter {
				return a, fmt.Errorf("unknown flag: %s\n%s", arg, usage)
			}
			args = append(args[:1], append(oneLetterFlags, args[1:]...)...)
			continue

		case len(arg) > 1 && arg[0] == '-':
			return a, fmt.Errorf("unknown flag: %s\n%s", arg, usage)

		default:
			rest = append(rest, arg)
			continue
		}
	}

	switch len(rest) {
	case 0:
	case 1:
		a.file = rest[0]
	default:
		return a, fmt.Errorf("too many file args\n%s", usage)
	}

	if a.bdump {
		switch a.bdumpFile {
		case "":
			if !strings.HasSuffix(a.file, ".bcl") {
				return a, fmt.Errorf(
					"--bdump requires knowing BFILE name, "+
						"either given as a flag, or derived from FILE"+
						"\n%s",
					usage,
				)
			}
			a.bdumpFile = a.file[:len(a.file)-len(".bcl")] + ".bcb"

		case "-":
			return a, fmt.Errorf("`-` is not a valid BFILE for --bdump")
		}
	}

	if a.bload {
		switch {
		case a.file == "" && a.bloadFile == "":
			// will use stdin; handled outside of this switch
		case a.file != "" && a.bloadFile != "":
			return a, fmt.Errorf("conflicting BFILE and FILE\n%s", usage)
		case a.file == "" && a.bloadFile != "":
			a.file = a.bloadFile
		}
	}

	if a.file == "" {
		a.file = "-"
	}
	return a, nil
}
