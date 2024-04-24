package main

import "fmt"

type parsedArgs struct {
	file   string
	disasm bool
	trace  bool
	mute   bool

	help func()
}

func parseArgs(args []string) (a parsedArgs, _ error) {
	const usage = "usage: bcl [-d|--disasm] [-t|--trace] [FILE|-]"

	for ; len(args) > 0; args = args[1:] {
		switch arg := args[0]; {
		case arg == "-h":
			a.help = func() { fmt.Println(usage) }
			return a, nil

		case arg == "-d" || arg == "--disasm":
			a.disasm = true
			continue

		case arg == "-t" || arg == "--trace":
			a.trace = true
			continue

		case arg == "--mute-result":
			a.mute = true
			continue

		case arg == "--":
			// todo: this should actually break flags processing
			continue

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
			if a.file != "" {
				return a, fmt.Errorf("too many file args\n%s", usage)
			}
			a.file = arg
			continue
		}
	}

	if a.file == "" {
		a.file = "-"
	}
	return a, nil
}
