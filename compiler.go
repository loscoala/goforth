package goforth

import (
	"fmt"
	"os"
	"strings"
)

// This is the public API for ForthCompiler.
type Compiler interface {
	Compile() error
	Parse(s string) error
	ParseFile(filename string) error
	StartREPL()
	RunFile(filename string) error
	Run(prog string) error
}

type ForthCompiler struct {
	label  Label
	blocks Label
	labels Stack[string]
	leaves Stack[string]
	whiles Stack[string]
	dos    Stack[string]
	cases  Stack[int]
	vars   Stack[string]
	funcs  map[string]*Stack[string]
	locals SliceStack[string]
	data   map[string]string
	defs   map[string]*Stack[string]
	output strings.Builder
	Fvm    *ForthVM
}

func NewForthCompiler() *ForthCompiler {
	fc := new(ForthCompiler)
	fc.data = map[string]string{
		"!":     "STR",
		"@":     "LV",
		".":     "PRI",
		"emit":  "PRA",
		"key":   "RDI",
		"=":     "EQI",
		"xor":   "XOR",
		"<":     "LSI",
		">":     "GRI",
		"-":     "SBI",
		"+":     "ADI",
		"/":     "DVI",
		"*":     "MLI",
		"f+":    "ADF",
		"f-":    "SBF",
		"f*":    "MLF",
		"f/":    "DVF",
		"f.":    "PRF",
		"f<":    "LSF",
		"f>":    "GRF",
		"not":   "NOT",
		"and":   "AND",
		"or":    "OR",
		"quit":  "STP",
		"dup":   "DUP",
		"2dup":  "TDP",
		"?dup":  "QDP",
		"over":  "OVR",
		"2over": "TVR",
		"drop":  "DRP",
		"swap":  "SWP",
		"2swap": "TWP",
		"sys":   "SYS",
		"rot":   "ROT",
		"exec":  "EXC",
		"pick":  "PCK",
		"-rot":  "NRT",
		">r":    "TR",
		"r>":    "FR",
		"r@":    "RF",
		"2>r":   "TTR",
		"2r>":   "TFR",
		"2r@":   "TRF",
		"inc":   "INC",
	}
	fc.funcs = make(map[string]*Stack[string])
	fc.defs = make(map[string]*Stack[string])
	fc.Fvm = NewForthVM()
	return fc
}

// The ByteCode of the "main" word previously compiled with Compile().
func (fc *ForthCompiler) ByteCode() string {
	return fc.output.String()
}

// Compiles the word "main".
// Side effect: ByteCode() contains the result.
func (fc *ForthCompiler) Compile() error {
	result := NewStack[string]()
	clear(fc.funcs)
	fc.output.Reset()

	if err := fc.compileWord("main", result); err != nil {
		return err
	}

	result.Push("STP")

	printVal := func(val string) {
		fc.output.WriteString(val)
		fc.output.WriteByte(';')
	}

	for _, v := range fc.funcs {
		v.Each(printVal)
	}

	fc.output.WriteString("MAIN;")
	result.Each(printVal)

	return nil
}

// Parses the given Forth code and adds the word to the dictionary of the compiler.
func (fc *ForthCompiler) Parse(str string) error {
	var (
		state   int
		counter int
		word    string
		def     *Stack[string]
	)

	buffer := make([]rune, 0, 100)

	for index, i := range str {
		switch state {
		case 0:
			switch i {
			case ':':
				state = 1
				def = NewStack[string]()
			case '\\':
				state = 4
			case '(':
				state = 5
			case '\n', '\r', '\t', ' ':
			default:
				state = 6
				buffer = append(buffer, i)
			}
		case 1:
			switch i {
			case '(':
				state = 3
			case '\\':
				state = 2
			case ';':
				fc.defs[word] = def
				counter = 0
				state = 0
			case '\n', '\r', '\t', ' ':
				if len(buffer) > 0 {
					if counter == 0 {
						word = string(buffer)
					} else {
						def.Push(string(buffer))
					}

					counter++
					buffer = buffer[:0]
				}
			case '.', 's':
				if index+1 == len(str) {
					break
				}

				if len(buffer) == 0 && str[index+1] == '"' {
					buffer = append(buffer, i)
					state = 7
				} else if len(buffer) == 0 && str[index+1] == '(' {
					buffer = append(buffer, i)
					state = 9
				} else {
					buffer = append(buffer, i)
				}
			default:
				buffer = append(buffer, i)
			}
		case 2:
			if i == '\n' {
				state = 1
			}
		case 3:
			if i == ')' {
				state = 1
			}
		case 4:
			if i == '\n' {
				state = 0
			}
		case 5:
			if i == ')' {
				state = 0
			}
		case 6:
			if i == '\n' {
				state = 0
				meta := string(buffer)
				if meta == "__END__" {
					return nil
				}
				if err := fc.handleMeta(meta); err != nil {
					return err
				}
				buffer = buffer[:0]
			} else if i != '\r' {
				buffer = append(buffer, i)
			}
		case 7:
			// consume "
			buffer = append(buffer, i)
			state = 8
		case 8:
			// inside string with .|s" "
			if index+1 == len(str) {
				break
			}

			buffer = append(buffer, i)

			if i == '\\' && str[index+1] == '"' {
				buffer = buffer[:len(buffer)-1]
				state = 7
			} else if i == '"' {
				handleForthString(def, buffer)
				buffer = buffer[:0]
				state = 1
			}
		case 9:
			// consume (
			buffer = append(buffer, i)
			state = 10
		case 10:
			// inside string with .|s( )
			if index+1 == len(str) {
				break
			}

			buffer = append(buffer, i)

			if i == '\\' && str[index+1] == ')' {
				buffer = buffer[:len(buffer)-1]
				state = 9
			} else if i == ')' {
				handleForthString(def, buffer)
				buffer = buffer[:0]
				state = 1
			}
		}
	}

	if state != 0 {
		return fmt.Errorf("syntax error: state should be 0 but is %d", state)
	}

	return nil
}

func compile_s(s *Stack[string], str []rune) {
	if len(str) > 9 {
		s.Push("0")

		for _, i := range reverse(str) {
			s.Push(fmt.Sprintf("%d", int(i)))
		}

		s.Push("print")
	} else {
		for _, i := range str {
			s.Push(fmt.Sprintf("%d", int(i)))
			s.Push("emit")
		}
	}
}

// modifies argument s to s' and returns s'
func reverse(s []rune) []rune {
	r := s // shallow copy

	// Optimization in order to omit make and copy
	// r := make([]rune, len(s))
	// copy(r, s)

	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return r
}

func compile_m(s *Stack[string], str []rune) {
	s.Push(">r")
	s.Push("0")

	for _, i := range reverse(str) {
		s.Push(fmt.Sprintf("%d", int(i)))
	}

	s.Push("r>")
	s.Push("!s")
}

func handleForthString(s *Stack[string], str []rune) {
	switch str[0] {
	case '.':
		compile_s(s, str[3:len(str)-1])
	case 's':
		compile_m(s, str[3:len(str)-1])
	}
}

func (fc *ForthCompiler) ParseFile(filename string) error {
	data, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	return fc.Parse(string(data))
}

func (fc *ForthCompiler) handleMeta(meta string) error {
	cmd := strings.Split(meta, " ")

	if cmd[0] == "use" {
		return fc.ParseFile(cmd[1])
	} else if cmd[0] == "variable" {
		if !fc.vars.Contains(cmd[1]) {
			fc.vars.Push(cmd[1])
		}
	}

	return nil
}

func (fc *ForthCompiler) compileLocals(iter *StackIter[string], result *Stack[string]) {
	localDefs := NewStack[string]()
	result.Push("LCTX")

	for iter.Next() {
		word := iter.Get()

		if word == "}" {
			break
		}

		localDefs.Push(word)
		result.Push("LDEF " + word)
	}

	fc.locals.Push(localDefs)
}

func (fc *ForthCompiler) compileBlock(iter *StackIter[string], result *Stack[string]) error {
	var blockCounter int

	blockName := fc.blocks.CreateNewWord()
	fc.defs[blockName] = NewStack[string]()

	for iter.Next() {
		word := iter.Get()

		if word == "[" {
			blockCounter++
		}

		if word == "]" {
			if blockCounter == 0 {
				break
			}

			blockCounter--
		}

		fc.defs[blockName].Push(word)
	}

	blockDef := NewStack[string]()
	blockDef.Push("SUB " + blockName)
	if err := fc.compileWordWithLocals(blockName, fc.defs[blockName], blockDef); err != nil {
		return err
	}
	blockDef.Push("END")
	fc.funcs[blockName] = blockDef
	result.Push("REF " + blockName)

	return nil
}

func (fc *ForthCompiler) compileWordWithLocals(word string, wordDef *Stack[string], result *Stack[string]) error {
	var localCounter int

	for iter := wordDef.Iter(); iter.Next(); {
		word2 := iter.Get()

		if word2 == "{" {
			localCounter++
			fc.compileLocals(iter, result)
		} else if word2 == "[" {
			if err := fc.compileBlock(iter, result); err != nil {
				return err
			}
		} else if word2 == "to" {
			iter.Next()
			word2 = iter.Get()
			if fc.vars.Contains(word2) {
				if _, ok := fc.funcs[word2]; !ok {
					gdef := NewStack[string]()
					gdef.Push("GDEF " + word2)
					fc.funcs[word2] = gdef
				}
				result.Push("GSET " + word2)
				continue
			}
			if !fc.locals.Contains(word2) {
				return fmt.Errorf("unable to assign word \"%s\": not in local context", word2)
			}
			result.Push("LSET " + word2)
		} else if word2 == "done" {
			localCounter--
			fc.locals.ExPop()
			result.Push("LCLR")
		} else if word2 == word {
			result.Push("CALL " + word)
		} else {
			if err := fc.compileWord(word2, result); err != nil {
				return err
			}
		}
	}

	for i := 0; i < localCounter; i++ {
		fc.locals.ExPop()
		result.Push("LCLR")
	}

	return nil
}

func isFloat(s string) bool {
	pos := strings.IndexByte(s, '.')

	if pos == -1 {
		return false
	}

	first := s[:pos]
	last := s[pos+1:]

	return (len(first) != 0 || len(last) != 0) && isNumeric(first) && isNumeric(last)
}

func isNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}

	return true
}

func (fc *ForthCompiler) compileWord(word string, result *Stack[string]) error {
	if isNumeric(word) {
		result.Push("L " + word)
	} else if isFloat(word) {
		result.Push("LF " + word)
	} else if fc.locals.Contains(word) {
		result.Push("LCL " + word)
	} else if fc.vars.Contains(word) {
		if _, ok := fc.funcs[word]; !ok {
			gdef := NewStack[string]()
			gdef.Push("GDEF " + word)
			fc.funcs[word] = gdef
		}
		result.Push("GBL " + word)
	} else if value, ok := fc.data[word]; ok {
		result.Push(value)
	} else if wordDef, ok := fc.defs[word]; ok {
		if word != "main" && wordDef.Len() > 4 {
			if _, ok := fc.funcs[word]; !ok {
				funcDef := NewStack[string]()
				funcDef.Push("SUB " + word)
				if err := fc.compileWordWithLocals(word, wordDef, funcDef); err != nil {
					return err
				}
				funcDef.Push("END")
				fc.funcs[word] = funcDef
			}

			result.Push("CALL " + word)
		} else {
			if err := fc.compileWordWithLocals(word, wordDef, result); err != nil {
				return err
			}
		}
	} else if word[0] == '&' {
		realWord := word[1:]
		if wordDef, ok := fc.defs[realWord]; ok {
			if _, ok := fc.funcs[realWord]; !ok {
				funcDef := NewStack[string]()
				funcDef.Push("SUB " + realWord)
				if err := fc.compileWordWithLocals(realWord, wordDef, funcDef); err != nil {
					return err
				}
				funcDef.Push("END")
				fc.funcs[realWord] = funcDef
			}
		} else {
			return fmt.Errorf("unable to reference word \"%s\": Unknown word", realWord)
		}
		result.Push("REF " + realWord)
	} else if word == "case" {
		fc.cases.Push(0)
	} else if word == "if" || word == "?of" || word == "of" {
		if word == "of" {
			result.Push("OVR")
			result.Push("EQI")
		}
		lbl := fc.label.CreateNewLabel()
		result.Push("JIN #" + lbl)
		fc.labels.Push(lbl)
		if word == "?of" || word == "of" {
			fc.cases.Push(fc.cases.ExPop() + 1)
		}
	} else if word == "else" || word == "endof" {
		lbl := fc.label.CreateNewLabel()
		result.Push("JMP #" + lbl)
		result.Push("#" + fc.labels.ExPop() + " NOP")
		fc.labels.Push(lbl)
	} else if word == "then" {
		result.Push("#" + fc.labels.ExPop() + " NOP")
	} else if word == "endcase" {
		num := fc.cases.ExPop()
		for i := 0; i < num; i++ {
			result.Push("#" + fc.labels.ExPop() + " NOP")
		}
	} else if word == "begin" {
		lbl := fc.label.CreateNewLabel()
		result.Push("#" + lbl + " NOP")
		fc.labels.Push(lbl)
	} else if word == "do" || word == "?do" {
		result.Push("TTR") // rstack: end i
		if word == "?do" {
			result.Push("TRF") // end i
			result.Push("EQI")
			result.Push("NOT")
			lbl := fc.label.CreateNewLabel()
			result.Push("JIN #" + lbl)
			fc.dos.Push(lbl)
		}
		lbl := fc.label.CreateNewLabel()
		result.Push("#" + lbl + " NOP")
		fc.labels.Push(lbl)
	} else if word == "while" {
		lbl := fc.label.CreateNewLabel()
		result.Push("JIN #" + lbl)
		fc.whiles.Push(lbl)
	} else if word == "loop" || word == "+loop" || word == "-loop" {
		result.Push("FR")
		if word == "-loop" {
			result.Push("SWP")
			result.Push("SBI")
		} else {
			if word == "loop" {
				result.Push("INC")
			} else {
				result.Push("ADI")
			}
		}
		result.Push("RF")  // i end
		result.Push("SWP") // end i
		result.Push("DUP") // end i i
		result.Push("TR")  // end i
		if word == "-loop" {
			result.Push("LSI")
		} else {
			result.Push("GRI")
		}
		result.Push("NOT")
		result.Push("JIN #" + fc.labels.ExPop())
		if fc.leaves.Len() > 0 {
			result.Push("#" + fc.leaves.ExPop() + " NOP")
		}
		if fc.dos.Len() > 0 {
			result.Push("#" + fc.dos.ExPop() + " NOP")
		}
		result.Push("TFR")
		result.Push("DRP")
		result.Push("DRP")
	} else if word == "leave" {
		lbl := fc.label.CreateNewLabel()
		result.Push("JMP #" + lbl)
		fc.leaves.Push(lbl)
	} else if word == "until" {
		result.Push("JIN #" + fc.labels.ExPop())
		if fc.leaves.Len() > 0 {
			result.Push("#" + fc.leaves.ExPop() + " NOP")
		}
	} else if word == "again" || word == "repeat" {
		result.Push("JMP #" + fc.labels.ExPop())
		if fc.leaves.Len() > 0 {
			result.Push("#" + fc.leaves.ExPop() + " NOP")
		}
		if word == "repeat" && fc.whiles.Len() > 0 {
			result.Push("#" + fc.whiles.ExPop() + " NOP")
		}
	} else {
		return fmt.Errorf("word \"%s\" unknown", word)
	}

	return nil
}
