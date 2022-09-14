package goforth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	HEADER    = "\033[95m"
	OKBLUE    = "\033[94m"
	OKCYAN    = "\033[96m"
	OKGREEN   = "\033[92m"
	WARNING   = "\033[93m"
	FAIL      = "\033[91m"
	ENDC      = "\033[0m"
	BOLD      = "\033[91m"
	UNDERLINE = "\033[4m"
)

var baseSyntax = [...]string{
	"begin", "while", "repeat", "do", "?do", "loop", "+loop", "-loop", "if", "then",
	"else", "{", "}", "[", "]", "until", "again", "leave", "to", "done", ":", ";",
}

func isBaseSytax(word string) bool {
	for _, w := range baseSyntax {
		if w == word {
			return true
		}
	}

	return false
}

func getWordColored(fc *ForthCompiler, word string) string {
	var color string

	if _, ok := fc.data[word]; ok {
		color = HEADER
	} else if _, ok := fc.defs[word]; ok {
		color = OKCYAN
	} else if isBaseSytax(word) {
		color = OKGREEN
	} else if isFloat(word) || isNumeric(word) {
		color = OKBLUE
	} else {
		color = WARNING
	}

	return fmt.Sprintf("%s%s%s", color, word, ENDC)
}

func printWordColored(fc *ForthCompiler, word string, s *Stack) {
	fmt.Printf("%s %s ", getWordColored(fc, ":"), getWordColored(fc, word))

	for iter := s.Iter(); iter.Next(); {
		fmt.Printf("%s ", getWordColored(fc, iter.Get()))
	}

	fmt.Printf("%s\n", getWordColored(fc, ";"))
}

func printWord(word string, s *Stack) {
	fmt.Printf(": %s ", word)

	for iter := s.Iter(); iter.Next(); {
		fmt.Printf("%s ", iter.Get())
	}

	fmt.Println(";")
}

func PrintError(err error) {
	if Colored {
		fmt.Printf("%s[Error]%s: %s\n", FAIL, ENDC, err)
	} else {
		fmt.Printf("[Error]: %s\n", err)
	}
}

func (fc *ForthCompiler) printDefinition(word string) {
	s, ok := fc.defs[word]

	if !ok {
		// primitive
		p, ok2 := fc.data[word]

		if !ok2 {
			if isBaseSytax(word) {
				if Colored {
					fmt.Printf("Word %s is a compiler builtin.\n", getWordColored(fc, word))
				} else {
					fmt.Printf("Word \"%s\" is a compiler builtin.\n", word)
				}
			} else {
				if Colored {
					fmt.Printf("Unknown word %s%s%s\n", FAIL, word, ENDC)
				} else {
					fmt.Printf("Unknown word \"%s\"\n", word)
				}
			}
		} else {
			if Colored {
				fmt.Printf("%s %s %s %s\n", getWordColored(fc, ":"), getWordColored(fc, word),
					getWordColored(fc, p), getWordColored(fc, ";"))
			} else {
				fmt.Printf(": %s %s ;\n", word, p)
			}
		}

		return
	}

	if Colored {
		printWordColored(fc, word, s)
	} else {
		printWord(word, s)
	}
}

func (fc *ForthCompiler) printAllDefinitions() {
	if Colored {
		for k, s := range fc.defs {
			printWordColored(fc, k, s)
		}
	} else {
		for k, s := range fc.defs {
			printWord(k, s)
		}
	}
	fmt.Println("")
}

func (fc *ForthCompiler) printByteCode() {
	if Colored {
		for _, cmd := range strings.Split(fc.ByteCode(), ";") {
			if cmd == "" {
				continue
			}
			fmt.Printf("%s%s%s;", WARNING, cmd, ENDC)
			if cmd == "END" {
				fmt.Println("")
			}
		}
		fmt.Println("")
	} else {
		fmt.Println(fc.ByteCode())
	}
}

func isWhiteSpace(s string) bool {
	for _, v := range s {
		if !unicode.IsSpace(v) {
			return false
		}
	}

	return true
}

func (fc *ForthCompiler) StartREPL() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("forth> ")
		scanner.Scan()
		text := scanner.Text()

		if isWhiteSpace(text) {
			continue
		}

		if text == "exit" {
			break
		}

		if text[0] == ':' {
			// just parse
			if err := fc.Parse(text); err != nil {
				PrintError(err)
			}
			continue
		}

		if text[0] == '%' && len(text) > 1 {
			// show just one definition
			fc.printDefinition(text[2:])
			continue
		} else if text[0] == '%' && len(text) == 1 {
			// show all definitions
			fc.printAllDefinitions()
			continue
		} else if text[0] == '#' && len(text) > 1 {
			// open a file an parse its contents
			if err := fc.ParseFile(text[2:]); err != nil {
				PrintError(err)
			}
			continue
		} else if text[0] == '$' && len(text) == 1 {
			for i := 0; i <= fc.Fvm.n; i++ {
				fmt.Printf("%d ", fc.Fvm.stack[i])
			}
			fmt.Println("")
			continue
		}

		if err := fc.Parse(": main\n" + text + "\n;"); err != nil {
			PrintError(err)
			continue
		}

		if err := fc.Compile(); err != nil {
			PrintError(err)
			continue
		}

		fc.printByteCode()

		fc.Fvm.Run(fc.ByteCode())
		fmt.Println("")
	}
}

func (fc *ForthCompiler) RunFile(str string) error {
	if err := fc.ParseFile(str); err != nil {
		return err
	}

	if err := fc.Compile(); err != nil {
		return err
	}

	fc.Fvm.Run(fc.ByteCode())

	return nil
}

func (fc *ForthCompiler) Run(prog string) error {
	if err := fc.Parse(prog); err != nil {
		return err
	}

	if err := fc.Compile(); err != nil {
		return err
	}

	fc.Fvm.Run(fc.ByteCode())

	return nil
}
