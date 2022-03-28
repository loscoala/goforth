package main

import (
	"os"
	"strings"
)

/* --------- For later use --------------
type Compiler interface {
	Compile(s string) *Stack
	Parse(s string)
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
	state := 0

	for _, i := range data {
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
		}
	}

	return string(result)
}

func (fc *ForthCompiler) ParseFile(filename string) {
	data, err := os.ReadFile(filename)

	if err != nil {
		panic(err)
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

func (fc *ForthCompiler) compileWordWithLocals(wordDef *Stack, result *Stack) {
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
				panic("NameError: Unable to assign word \"" + word2 + "\": not in local context")
			}
			result.Push("LSET " + word2)
			assignMode = false
		} else {
			fc.compileWord(word2, result)
		}
	}
	for i := 0; i < localCounter; i++ {
		fc.locals.ExPop()
		result.Push("LCLR")
	}
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
	} else if value, ok := fc.data[word]; ok {
		result.Push(value)
	} else if wordDef, ok := fc.defs[word]; ok {
		if word != "main" && wordDef.Len() > 4 {
			if _, ok := fc.funcs[word]; !ok {
				funcDef := new(Stack)
				funcDef.Push("SUB " + word)
				fc.compileWordWithLocals(wordDef, funcDef)
				funcDef.Push("END")
				fc.funcs[word] = funcDef
			}

			result.Push("CALL " + word)
		} else {
			fc.compileWordWithLocals(wordDef, result)
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
	} else if word == "loop" || word == "+loop" {
		if word == "loop" {
			result.Push("L 1")
		}
		result.Push("LCL i")
		result.Push("ADI")
		result.Push("LSET i")
		result.Push("LCL end")
		result.Push("LCL i")
		result.Push("GRI")
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
		panic("Word \"" + word + "\" unknown")
	}
}
