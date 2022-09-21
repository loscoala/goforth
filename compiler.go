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
	labels Stack
	leaves Stack
	whiles Stack
	dos    Stack
	funcs  map[string]*Stack
	locals SliceStack
	data   map[string]string
	defs   map[string]*Stack
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
	}
	fc.defs = make(map[string]*Stack)
	fc.Fvm = NewForthVM()
	return fc
}

func (fc *ForthCompiler) ByteCode() string {
	return fc.output.String()
}

func (fc *ForthCompiler) Compile() error {
	result := NewStack()
	fc.funcs = make(map[string]*Stack)
	fc.output.Reset()

	if err := fc.compileWord("main", result); err != nil {
		return err
	}

	result.Push("STP")

	for _, v := range fc.funcs {
		for iter := v.Iter(); iter.Next(); {
			fc.output.WriteString(iter.Get())
			fc.output.WriteByte(';')
		}
	}

	fc.output.WriteString("MAIN;")

	for iter := result.Iter(); iter.Next(); {
		fc.output.WriteString(iter.Get())
		fc.output.WriteByte(';')
	}

	return nil
}

func (fc *ForthCompiler) parseAuto(data string) (string, error) {
	result := make([]rune, 0, len(data)+1)
	tmpStr := make([]rune, 0, 100)
	state := 0

	for index, i := range data {
		switch state {
		case 0:
			switch i {
			case ':':
				state = 1
				result = append(result, i)
			case '\\':
				state = 4
			case '(':
				state = 5
			case '\n', '\r', '\t', ' ':
			default:
				state = 6
				tmpStr = append(tmpStr, i)
			}
		case 1:
			switch i {
			case '(':
				state = 3
			case '\\':
				state = 2
			case ';':
				state = 0
				result = append(result, i)
				result = append(result, ' ')
			case '\n', '\r', '\t':
				result = append(result, ' ')
			case '.':
				if index+1 == len(data) {
					break
				}
				if data[index+1] == '"' {
					tmpStr = append(tmpStr, i)
					state = 7
				} else {
					result = append(result, i)
				}
			case 's':
				if index+1 == len(data) {
					break
				}
				if data[index+1] == '"' {
					tmpStr = append(tmpStr, i)
					state = 7
				} else {
					result = append(result, i)
				}
			default:
				result = append(result, i)
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
				if err := fc.handleMeta(string(tmpStr)); err != nil {
					return "", err
				}
				tmpStr = tmpStr[:0]
			} else if i != '\r' {
				tmpStr = append(tmpStr, i)
			}
		case 7:
			// consume "
			tmpStr = append(tmpStr, i)
			state = 8
		case 8:
			// inside string
			if index+1 == len(data) {
				break
			}

			tmpStr = append(tmpStr, i)

			if i == '\\' && data[index+1] == '"' {
				tmpStr = tmpStr[:len(tmpStr)-1]
				state = 7
			} else if i == '"' {
				result = append(result, handleForthString(string(tmpStr))...)
				tmpStr = tmpStr[:0]
				state = 1
			}
		}
	}

	if state != 0 {
		return "", fmt.Errorf("syntax error: state should be 0 but is %d", state)
	}

	// shortCode := regexp.MustCompile(`\s+`).ReplaceAllString(string(result), " ")
	// return shortCode, nil

	return string(result), nil
}

func compile_s(str string) []rune {
	result := make([]rune, 0, 100)

	if len(str) > 9 {
		result = append(result, []rune("0 ")...)

		for _, i := range reverse(str) {
			result = append(result, []rune(fmt.Sprintf("%d ", int(i)))...)
		}

		result = append(result, []rune("print ")...)
	} else {
		for _, i := range str {
			result = append(result, []rune(fmt.Sprintf("%d emit ", int(i)))...)
		}
	}

	return result
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func compile_m(str string) []rune {
	result := make([]rune, 0, 100)
	result = append(result, []rune(">r 0 ")...)

	for _, i := range reverse(str) {
		result = append(result, []rune(fmt.Sprintf("%d ", int(i)))...)
	}

	result = append(result, []rune("r> !s ")...)

	return result
}

func handleForthString(str string) []rune {
	fstring := strings.Split(str, " ")

	switch fstring[0] {
	case ".\"":
		return compile_s(str[3 : len(str)-1])
	case "s\"":
		return compile_m(str[3 : len(str)-1])
	}

	return nil
}

func (fc *ForthCompiler) ParseFile(filename string) error {
	data, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	return fc.Parse(string(data))
}

func (fc *ForthCompiler) Parse(str string) error {
	var (
		inside bool
		first  bool
		word   string
	)

	code, err := fc.parseAuto(str)

	if err != nil {
		return err
	}

	for _, i := range strings.Split(code, " ") {
		if i == ":" {
			first = true
		} else if i == ";" {
			inside = false
		} else if i == "" {
			continue
		} else if first {
			fc.defs[i] = NewStack()
			word = i
			first = false
			inside = true
		} else if inside {
			fc.defs[word].Push(i)
		}
	}

	return nil
}

func (fc *ForthCompiler) handleMeta(meta string) error {
	cmd := strings.Split(meta, " ")

	if cmd[0] == "use" {
		return fc.ParseFile(cmd[1])
	}

	return nil
}

func (fc *ForthCompiler) compileLocals(iter *StackIter, result *Stack) {
	localDefs := NewStack()
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

func (fc *ForthCompiler) compileBlock(iter *StackIter, result *Stack) error {
	var blockCounter int

	blockName := fc.blocks.CreateNewWord()
	fc.defs[blockName] = NewStack()

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

	blockDef := NewStack()
	blockDef.Push("SUB " + blockName)
	if err := fc.compileWordWithLocals(blockName, fc.defs[blockName], blockDef); err != nil {
		return err
	}
	blockDef.Push("END")
	fc.funcs[blockName] = blockDef
	result.Push("REF " + blockName)

	return nil
}

func (fc *ForthCompiler) compileWordWithLocals(word string, wordDef *Stack, result *Stack) error {
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

func (fc *ForthCompiler) compileWord(word string, result *Stack) error {
	if isNumeric(word) {
		result.Push("L " + word)
	} else if isFloat(word) {
		result.Push("LF " + word)
	} else if fc.locals.Len() > 0 && fc.locals.Contains(word) {
		result.Push("LCL " + word)
	} else if value, ok := fc.data[word]; ok {
		result.Push(value)
	} else if wordDef, ok := fc.defs[word]; ok {
		if word != "main" && wordDef.Len() > 4 {
			if _, ok := fc.funcs[word]; !ok {
				funcDef := NewStack()
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
				funcDef := NewStack()
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
	} else if word == "if" {
		lbl := fc.label.CreateNewLabel()
		result.Push("JIN #" + lbl)
		fc.labels.Push(lbl)
	} else if word == "else" {
		lbl := fc.label.CreateNewLabel()
		result.Push("JMP #" + lbl)
		result.Push("#" + fc.labels.ExPop() + " NOP")
		fc.labels.Push(lbl)
	} else if word == "then" {
		result.Push("#" + fc.labels.ExPop() + " NOP")
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
				result.Push("L 1")
			}
			result.Push("ADI")
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
