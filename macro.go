package goforth

import (
	"fmt"
	"strconv"
	"strings"
)

type MacroCompiler struct {
	label  Label
	labels Stack[string]
	whiles Stack[string]
}

type MacroOptcode int

const (
	M_L MacroOptcode = iota
	M_STR
	M_NUM_ARGS
	M_DEPTH
	M_PUSH
	M_GRI
	M_LSI
	M_EQI
	M_NOT
	M_ADI
	M_PRINT_STACK
	M_JIN
	M_JMP
	M_NOP
	M_DUP
	M_DROP
	M_SWAP
	M_PRS
	M_PRINT
	M_STP
)

var MacroName = map[MacroOptcode]string{
	M_L:           "L",
	M_NUM_ARGS:    "NUM_ARGS",
	M_DEPTH:       "DEPTH",
	M_PUSH:        "PUSH",
	M_GRI:         "GRI",
	M_EQI:         "EQI",
	M_NOT:         "NOT",
	M_ADI:         "ADI",
	M_LSI:         "LSI",
	M_PRINT_STACK: "PRINT_STACK",
	M_JIN:         "JIN",
	M_JMP:         "JMP",
	M_NOP:         "NOP",
	M_DUP:         "DUP",
	M_DROP:        "DROP",
	M_SWAP:        "SWAP",
	M_PRS:         "PRS",
	M_PRINT:       "PRINT",
	M_STP:         "STP",
}

type Mc struct {
	cmd    MacroOptcode
	arg    string
	argInt int
}

func (mc *MacroCompiler) Compile(macroDef *Stack[string]) *Stack[*Mc] {
	r := NewStack[*Mc]()

	for macroWord := range macroDef.Values() {
		length := len(macroWord)

		if length > 2 && (macroWord[0] == '@' && macroWord[length-1] == '@') {
			inner := macroWord[1 : length-1]
			r.Push(&Mc{cmd: M_L, arg: inner})
			continue
		}

		if length > 2 && (macroWord[0] == '#' && macroWord[length-1] == '#') {
			inner := macroWord[1 : length-1]
			r.Push(&Mc{cmd: M_STR, arg: inner})
			continue
		}

		switch macroWord {
		case "@numArgs":
			r.Push(&Mc{cmd: M_NUM_ARGS})
		case "@depth":
			r.Push(&Mc{cmd: M_DEPTH})
		case "@push":
			r.Push(&Mc{cmd: M_PUSH})
		case "@>":
			r.Push(&Mc{cmd: M_GRI})
		case "@<":
			r.Push(&Mc{cmd: M_LSI})
		case "@=":
			r.Push(&Mc{cmd: M_EQI})
		case "@not":
			r.Push(&Mc{cmd: M_NOT})
		case "@$":
			r.Push(&Mc{cmd: M_PRINT_STACK})
		case "@dup":
			r.Push(&Mc{cmd: M_DUP})
		case "@drop":
			r.Push(&Mc{cmd: M_DROP})
		case "@swap":
			r.Push(&Mc{cmd: M_SWAP})
		case "@.":
			r.Push(&Mc{cmd: M_PRS})
		case "@add":
			r.Push(&Mc{cmd: M_ADI})
		case "@if":
			lbl := mc.label.CreateNewLabel()
			r.Push(&Mc{cmd: M_JIN, arg: lbl})
			mc.labels.Push(lbl)
		case "@else":
			lbl := mc.label.CreateNewLabel()
			r.Push(&Mc{cmd: M_JMP, arg: lbl})
			r.Push(&Mc{cmd: M_NOP, arg: mc.labels.ExPop()})
			mc.labels.Push(lbl)
		case "@then":
			r.Push(&Mc{cmd: M_NOP, arg: mc.labels.ExPop()})
		case "@begin":
			lbl := mc.label.CreateNewLabel()
			r.Push(&Mc{cmd: M_NOP, arg: lbl})
			mc.labels.Push(lbl)
		case "@while":
			lbl := mc.label.CreateNewLabel()
			r.Push(&Mc{cmd: M_JIN, arg: lbl})
			mc.whiles.Push(lbl)
		case "@repeat":
			r.Push(&Mc{cmd: M_JMP, arg: mc.labels.ExPop()})
			r.Push(&Mc{cmd: M_NOP, arg: mc.whiles.ExPop()})
		default:
			r.Push(&Mc{cmd: M_PRINT, arg: macroWord})
		}
	}

	r.Push(&Mc{cmd: M_STP})

	// optimise JIN and JMP
	for index, nop := range r.All() {
		if nop.cmd == M_NOP {
			for c := range r.Values() {
				if (c.cmd == M_JIN || c.cmd == M_JMP) && nop.arg == c.arg {
					c.argInt = index
				}
			}
		}
	}

	return r
}

type MacroVM struct {
	register map[string]*Stack[string]
	stack    *Stack[string]
}

func NewMacroVM() *MacroVM {
	r := new(MacroVM)
	r.register = make(map[string]*Stack[string])
	r.stack = NewStack[string]()
	return r
}

func (vm *MacroVM) wordInRegister(wordDef *Stack[string], register string) (*Stack[string], error) {
	var (
		word  string
		ok    bool
		count int
	)

	if word, ok = wordDef.Pop(); !ok {
		return nil, fmt.Errorf("unable to pop \"%s\". Not enough arguments", register)
	}

	result := NewStack[string]()

	if word == "]" {
		// inside block
		count = 1
		for {
			if word, ok = wordDef.Pop(); !ok {
				return nil, fmt.Errorf("unable to pop word from block definition. Number of \"]\" and of \"[\" is not equal")
			}
			if word == "[" {
				count--
				if count == 0 {
					break
				}
			} else if word == "]" {
				count++
			}

			result.Push(word)
		}

		result.Reverse()
	} else {
		// single word
		result.Push(word)
	}

	return result, nil
}

func numberOfBlocksOrWords(data *Stack[string]) int {
	var (
		counter, result int
	)

	for _, word := range data.Backward() {
		switch word {
		case "]":
			counter++
		case "[":
			counter--
			if counter == 0 {
				result++
			}
		default:
			if counter == 0 {
				result++
			}
		}
	}

	return result
}

func popToInt(s *Stack[string]) (int64, error) {
	a := s.ExPop()

	if !isNumeric(a) {
		return 0, fmt.Errorf("unable to parse %s as integer", a)
	}

	if b, err := strconv.ParseInt(a, 10, 64); err != nil {
		return 0, err
	} else {
		return b, nil
	}
}

func (vm *MacroVM) Run(code *Stack[*Mc], result *Stack[string]) error {
	done := false
	defer clear(vm.register)

	for progPtr := 0; !done; progPtr++ {
		cmd := code.data[progPtr]

		switch cmd.cmd {
		case M_L:
			var err error
			if vm.register[cmd.arg], err = vm.wordInRegister(result, cmd.arg); err != nil {
				return err
			}
		case M_STR:
			for w := range vm.register[cmd.arg].Values() {
				result.Push(w)
			}
		case M_NUM_ARGS:
			vm.stack.Push(fmt.Sprint(numberOfBlocksOrWords(result)))
		case M_DEPTH:
			vm.stack.Push(fmt.Sprint(vm.stack.Len()))
		case M_PUSH:
			var (
				err  error
				word *Stack[string]
			)
			if word, err = vm.wordInRegister(result, cmd.arg); err != nil {
				return err
			}
			for _, w := range word.Backward() {
				vm.stack.Push(w)
			}
		case M_GRI:
			var (
				a, b int64
				err  error
			)
			if a, err = popToInt(vm.stack); err != nil {
				return err
			}
			if b, err = popToInt(vm.stack); err != nil {
				return err
			}
			if a < b {
				vm.stack.Push("1")
			} else {
				vm.stack.Push("0")
			}
		case M_LSI:
			var (
				a, b int64
				err  error
			)
			if a, err = popToInt(vm.stack); err != nil {
				return err
			}
			if b, err = popToInt(vm.stack); err != nil {
				return err
			}
			if a > b {
				vm.stack.Push("1")
			} else {
				vm.stack.Push("0")
			}
		case M_EQI:
			a := vm.stack.ExPop()
			b := vm.stack.ExPop()

			if a == b {
				vm.stack.Push("1")
			} else {
				vm.stack.Push("0")
			}
		case M_NOT:
			a := vm.stack.ExPop()

			if a == "0" {
				vm.stack.Push("1")
			} else {
				vm.stack.Push("0")
			}
		case M_ADI:
			var (
				a, b int64
				err  error
			)
			if a, err = popToInt(vm.stack); err != nil {
				return err
			}
			if b, err = popToInt(vm.stack); err != nil {
				return err
			}
			vm.stack.Push(fmt.Sprint(a + b))
		case M_PRINT_STACK:
			for _, word := range vm.stack.Backward() {
				result.Push(word)
			}
			vm.stack.Reset()
		case M_JIN:
			a := vm.stack.ExPop()
			if a == "0" {
				progPtr = cmd.argInt
			}
		case M_JMP:
			progPtr = cmd.argInt
		case M_NOP:
			// pass
		case M_DUP:
			vm.stack.Push(vm.stack.data[len(vm.stack.data)-1])
		case M_DROP:
			vm.stack.ExPop()
		case M_SWAP:
			n := len(vm.stack.data) - 1
			a := vm.stack.data[n]
			vm.stack.data[n] = vm.stack.data[n-1]
			vm.stack.data[n-1] = a
		case M_PRS:
			result.Push(vm.stack.ExPop())
		case M_PRINT:
			arg := cmd.arg
			if isString(arg) {
				for key := range vm.register {
					marker := fmt.Sprintf("#%s#", key)

					if strings.Contains(arg, marker) {
						def := strings.Join(vm.register[key].data, " ")
						arg = strings.ReplaceAll(arg, marker, def)
					}
				}
			}
			result.Push(arg)
		case M_STP:
			done = true
		default:
			return fmt.Errorf("unknown opcode")
		}
	}

	return nil
}
