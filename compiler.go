package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

/* --------- For later use --------------
type Compiler interface {
	Compile()
	Parse(s string)
	ParseFile(filename string)
	StartREPL()
	RunFile(filename string)
}
*/

type ForthCompiler struct {
	label  Label
	labels Stack
	leaves Stack
	whiles Stack
	dos    Stack
	funcs  map[string]*Stack
	locals SliceStack
	data   map[string]string
	defs   map[string]*Stack
	output strings.Builder
}

func NewForthCompiler() *ForthCompiler {
	ft := new(ForthCompiler)
	ft.data = map[string]string{
		"!":     "STR",
		"@":     "LV",
		".":     "PRI",
		"emit":  "PRA",
		"key":   "RDI",
		"=":     "EQI",
		":":     "",
		";":     "",
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
	}
	ft.defs = make(map[string]*Stack)
	return ft
}

func (fc *ForthCompiler) ByteCode() string {
	return fc.output.String()
}

func (fc *ForthCompiler) Compile() {
	result := new(Stack)
	fc.funcs = make(map[string]*Stack)
	fc.output.Reset()

	fc.compileWord("main", result)
	result.Push("STP")

	for _, v := range fc.funcs {
		for iter := v.Iter(); iter.Next(); {
			fc.output.WriteString(iter.Get() + ";")
		}
	}

	fc.output.WriteString("MAIN;")

	for iter := result.Iter(); iter.Next(); {
		fc.output.WriteString(iter.Get() + ";")
	}
}

func parseAuto(data string) string {
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
			case '\n':
			case '\t':
			case ' ':
			default:
				state = 6
				// state6_line.append
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
			case '\n':
				result = append(result, ' ')
			case '\t':
				result = append(result, ' ')
			case '.':
				if data[index+1] == '"' {
					tmpStr = append(tmpStr, i)
					state = 7
				} else {
					result = append(result, i)
				}
			case '!':
				if data[index+1] == '"' {
					tmpStr = append(tmpStr, i)
					state = 7
				} else {
					result = append(result, i)
				}
			case 's':
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
				// TODO:
			} // else {
			// TODO:
			// }
		case 7:
			// consume "
			tmpStr = append(tmpStr, i)
			state = 8
		case 8:
			// inside string
			tmpStr = append(tmpStr, i)

			if i == '"' {
				result = append(result, handleForthString(string(tmpStr))...)
				tmpStr = tmpStr[:0]
				state = 1
			}
		}
	}

	return string(result)
}

func compile_s(str string) []rune {
	result := make([]rune, 0, 100)

	for _, i := range str {
		result = append(result, []rune(fmt.Sprintf("%d emit ", int(i)))...)
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

func compile_m(str string, base int) []rune {
	result := make([]rune, 0, 100)
	result = append(result, []rune("0 ")...)

	for _, i := range reverse(str) {
		result = append(result, []rune(fmt.Sprintf("%d ", int(i)))...)
	}

	result = append(result, []rune(fmt.Sprintf("%d !s ", base))...)

	return result
}

func compile_m2(str string) []rune {
	result := make([]rune, 0, 100)
	result = append(result, []rune("{ n } 0 ")...)

	for _, i := range reverse(str) {
		result = append(result, []rune(fmt.Sprintf("%d ", int(i)))...)
	}

	result = append(result, []rune("n !s end ")...)

	return result
}

func handleForthString(str string) []rune {
	fstring := strings.Split(str, " ")

	switch fstring[0] {
	case ".\"":
		return compile_s(str[3 : len(str)-1])
	case "!\"":
		base, err := strconv.Atoi(fstring[1])
		if err != nil {
			log.Fatal(err)
		}
		pos := 4 + len(fstring[1])
		return compile_m(str[pos:len(str)-1], base)
	case "s\"":
		return compile_m2(str[3 : len(str)-1])
	default:
		log.Fatalf("Unknown type of string found %s\n", fstring[0])
	}

	return nil
}

func (fc *ForthCompiler) ParseFile(filename string) {
	data, err := os.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	fc.Parse(string(data))
}

// {"main": ["1", "1", "+", ".", "quit"], "add": ["+"]}
func (fc *ForthCompiler) Parse(str string) {
	var inside bool
	var first bool
	var word string

	for _, i := range strings.Split(parseAuto(str), " ") {
		if i == ":" {
			first = true
		} else if i == ";" {
			inside = false
		} else if i == "" {
			continue
		} else if first {
			fc.defs[i] = new(Stack)
			word = i
			first = false
			inside = true
		} else if inside {
			fc.defs[word].Push(i)
		}
	}
}

func (fc *ForthCompiler) compileWordWithLocals(word string, wordDef *Stack, result *Stack) {
	localMode := false
	localCounter := 0
	var localDefs *Stack
	assignMode := false

	for iter := wordDef.Iter(); iter.Next(); {
		word2 := iter.Get()

		if word2 == "{" {
			localMode = true
			localCounter++
			localDefs = new(Stack)
			result.Push("LCTX")
			continue
		}

		if localMode {
			if word2 == "}" {
				localMode = false
				fc.locals.Push(localDefs)
			} else {
				localDefs.Push(word2)
				result.Push("LDEF " + word2)
			}
		} else if word2 == "to" {
			assignMode = true
		} else if assignMode {
			if !fc.locals.Contains(word2) {
				log.Fatal("NameError: Unable to assign word \"" + word2 + "\": not in local context")
			}
			result.Push("LSET " + word2)
			assignMode = false
		} else if word2 == word {
			result.Push("CALL " + word)
		} else if word2 == "end" {
			localCounter--
			fc.locals.ExPop()
			result.Push("LCLR")
		} else {
			fc.compileWord(word2, result)
		}
	}
	for i := 0; i < localCounter; i++ {
		fc.locals.ExPop()
		result.Push("LCLR")
	}
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

func (fc *ForthCompiler) compileWord(word string, result *Stack) {
	if isNumeric(word) {
		result.Push("L " + word)
	} else if isFloat(word) {
		result.Push("LF " + word)
	} else if value, ok := fc.data[word]; ok {
		result.Push(value)
	} else if wordDef, ok := fc.defs[word]; ok {
		if word != "main" && wordDef.Len() > 4 {
			if _, ok := fc.funcs[word]; !ok {
				funcDef := new(Stack)
				funcDef.Push("SUB " + word)
				fc.compileWordWithLocals(word, wordDef, funcDef)
				funcDef.Push("END")
				fc.funcs[word] = funcDef
			}

			result.Push("CALL " + word)
		} else {
			fc.compileWordWithLocals(word, wordDef, result)
		}
	} else if fc.locals.Len() > 0 && fc.locals.Contains(word) {
		result.Push("LCL " + word)
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
		result.Push("LCTX")
		result.Push("LDEF i")
		result.Push("LDEF end")
		{
			stk := new(Stack)
			stk.Push("i")
			stk.Push("end")
			fc.locals.Push(stk)
		}
		if word == "?do" {
			result.Push("LCL end")
			result.Push("LCL i")
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
		if word == "-loop" {
			result.Push("LCL i")
			result.Push("SWP")
			result.Push("SBI")
		} else {
			if word == "loop" {
				result.Push("L 1")
			}
			result.Push("LCL i")
			result.Push("ADI")
		}
		result.Push("LSET i")
		result.Push("LCL end")
		result.Push("LCL i")
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
		fc.locals.ExPop()
		result.Push("LCLR")
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
		log.Fatal("Word \"" + word + "\" unknown")
	}
}
