package goforth

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	label         Label
	blocks        Label
	labels        Stack[string]
	leaves        Stack[string]
	whiles        Stack[string]
	dos           Stack[string]
	cases         Stack[int]
	vars          Stack[string]
	funcs         map[string]*Stack[string]
	locals        SliceStack[string]
	data          map[string]string
	defs          map[string]*Stack[string]
	inlines       map[string]*Stack[string]
	macroRegister [4]Stack[string]
	output        strings.Builder
	Fvm           *ForthVM
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
		"dec":   "DEC",
	}
	fc.funcs = make(map[string]*Stack[string])
	fc.defs = make(map[string]*Stack[string])
	fc.inlines = make(map[string]*Stack[string])
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
	fc.label.Reset()
	fc.blocks.Reset()

	if err := fc.compileWord("main", result); err != nil {
		return err
	}

	result.Push("L 0")
	result.Push("STP")

	printVal := func(val string) {
		fc.output.WriteString(val)
		fc.output.WriteByte(';')
	}

	for _, v := range fc.funcs {
		for val := range v.All() {
			printVal(val)
		}
	}

	fc.output.WriteString("MAIN;")

	for val := range result.All() {
		printVal(val)
	}

	return nil
}

// Parses the given Forth code and adds the word to the dictionary of the compiler.
func (fc *ForthCompiler) Parse(str, filename string) error {
	var (
		state   int
		counter int
		word    string
		def     *Stack[string]
	)

	buffer := make([]rune, 0, 100)
	line := 1
	pos := 0

	for index, i := range str {
		pos++
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
			case '\r', '\t', ' ':
			case '\n':
				line++
				pos = 1
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
				switch word {
				case "class":
					if err := fc.compileClass(def, filename); err != nil {
						return fmt.Errorf("%s Line %d at %d: %s", filename, line, pos, err.Error())
					}
				case "inline":
					word = def.data[0]
					if _, ok := fc.defs[word]; ok {
						return fmt.Errorf("unable to define inline. \"%s\" is already defined as word", word)
					}
					tmp := new(Stack[string])
					tmp.data = def.data[1:]
					fc.inlines[word] = tmp
				default:
					if _, ok := fc.inlines[word]; ok {
						return fmt.Errorf("unable to define word. \"%s\" is already defined as inline", word)
					}
					fc.defs[word] = def
				}

				counter = 0
				state = 0
			case '\n', '\r', '\t', ' ':
				if i == '\n' {
					line++
					pos = 1
				}
				if len(buffer) > 0 {
					if counter == 0 {
						word = string(buffer)
					} else {
						tmp := string(buffer)
						if inline, ok := fc.inlines[tmp]; ok {
							switch inline.data[0] {
							case "@1@":
								if err := fc.wordInRegister(def, 0); err != nil {
									return err
								}
							case "@2@":
								if err := fc.wordInRegister(def, 0); err != nil {
									return err
								}
								if err := fc.wordInRegister(def, 1); err != nil {
									return err
								}
							case "@3@":
								if err := fc.wordInRegister(def, 0); err != nil {
									return err
								}
								if err := fc.wordInRegister(def, 1); err != nil {
									return err
								}
								if err := fc.wordInRegister(def, 2); err != nil {
									return err
								}
							case "@4@":
								if err := fc.wordInRegister(def, 0); err != nil {
									return err
								}
								if err := fc.wordInRegister(def, 1); err != nil {
									return err
								}
								if err := fc.wordInRegister(def, 2); err != nil {
									return err
								}
								if err := fc.wordInRegister(def, 3); err != nil {
									return err
								}
							default:
								// skip
							}

							for value := range inline.All() {
								pushToDef := func(val string) {
									def.Push(val)
								}
								switch value {
								case "#1#":
									for val := range fc.macroRegister[0].All() {
										pushToDef(val)
									}
								case "#2#":
									for val := range fc.macroRegister[1].All() {
										pushToDef(val)
									}
								case "#3#":
									for val := range fc.macroRegister[2].All() {
										pushToDef(val)
									}
								case "#4#":
									for val := range fc.macroRegister[3].All() {
										pushToDef(val)
									}
								case "@1@", "@2@", "@3@", "@4@":
									// skip
								default:
									if isString(value) {
										if strings.Contains(value, "#1#") {
											d := strings.Join(fc.macroRegister[0].data, " ")
											def.Push(strings.ReplaceAll(value, "#1#", d))
										} else if strings.Contains(value, "#2#") {
											d := strings.Join(fc.macroRegister[1].data, " ")
											def.Push(strings.ReplaceAll(value, "#2#", d))
										} else if strings.Contains(value, "#3#") {
											d := strings.Join(fc.macroRegister[2].data, " ")
											def.Push(strings.ReplaceAll(value, "#3#", d))
										} else if strings.Contains(value, "#4#") {
											d := strings.Join(fc.macroRegister[3].data, " ")
											def.Push(strings.ReplaceAll(value, "#4#", d))
										}
									} else {
										def.Push(value)
									}
								}
							}

							// clean all registers
							for i := range 4 {
								fc.macroRegister[i].Reset()
							}
						} else {
							def.Push(tmp)
						}
					}

					counter++
					buffer = buffer[:0]
				}
			case '.', 'a', 'g':
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
				line++
				pos = 1
			}
		case 3:
			if i == ')' {
				state = 1
			}
		case 4:
			if i == '\n' {
				state = 0
				line++
				pos = 1
			}
		case 5:
			if i == ')' {
				state = 0
			}
		case 6:
			if i == '\n' {
				state = 0
				pos = 1
				meta := string(buffer)
				if meta == "__END__" {
					return nil
				}
				if err := fc.handleMeta(meta); err != nil {
					return fmt.Errorf("%s Line %d at %d: %s", filename, line, pos, err.Error())
				}
				line++
				buffer = buffer[:0]
			} else if i != '\r' {
				buffer = append(buffer, i)
			}
		case 7:
			// consume "
			buffer = append(buffer, i)
			state = 8
		case 8:
			// inside string with .|a"|g" "
			if index+1 == len(str) {
				break
			}

			buffer = append(buffer, i)

			if i == '\\' && str[index+1] == '"' {
				buffer = buffer[:len(buffer)-1]
				state = 7
			} else if i == '"' {
				def.Push(string(buffer))
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
				def.Push(string(buffer))
				buffer = buffer[:0]
				state = 1
			}
		}
	}

	if state != 0 {
		var cause string
		switch state {
		case 1:
			cause = "word definition is not closed"
		case 8:
			cause = "missing '\"'"
		case 3, 5, 10:
			cause = "missing ')'"
		default:
			cause = fmt.Sprintf("parser state is %d but should be 0", state)
		}
		return fmt.Errorf("%s Line %d at %d: syntax error: %s", filename, line, pos, cause)
	}

	return nil
}

// inline block or single word
func (fc *ForthCompiler) wordInRegister(wordDef *Stack[string], index int) error {
	var (
		word  string
		ok    bool
		count int
	)

	if word, ok = wordDef.Pop(); !ok {
		return fmt.Errorf("unable to pop %dth word from word definition. Not enough arguments", index+1)
	}

	if word == "]" {
		// inside block
		count = 1
		for {
			if word, ok = wordDef.Pop(); !ok {
				return fmt.Errorf("unable to pop word from block definition. Number of \"]\" and of \"[\" is not equal")
			}
			if word == "[" {
				count--
				if count == 0 {
					break
				}
			} else if word == "]" {
				count++
			}

			fc.macroRegister[index].Push(word)
		}
		fc.macroRegister[index].Reverse()
	} else {
		// single word
		fc.macroRegister[index].Push(word)
	}

	return nil
}

func (fc *ForthCompiler) ParseTemplate(entry, str, filename string) error {
	var (
		state  int
		buffer strings.Builder
	)

	buffer.Grow(len(str) + len(entry) + 50)
	buffer.WriteString(": ")
	buffer.WriteString(entry)
	buffer.WriteString(" g( ")

	gEnd := fmt.Sprintf(") %s:print", entry)

	for i := 0; i < len(str); i++ {
		switch state {
		case 0:
			if str[i] == '<' &&
				str[i+1] == '?' &&
				str[i+2] == 'f' &&
				str[i+3] == 's' {
				i += 3
				state = 1
				buffer.WriteString(gEnd)
				continue
			}

			if str[i] == ')' {
				buffer.WriteByte('\\')
			}

			buffer.WriteByte(str[i])
		case 1:
			if str[i] == '?' &&
				str[i+1] == '>' {
				i += 1
				if str[i+1] == '\n' {
					i += 1
				}
				state = 0
				buffer.WriteString("g( ")
				continue
			}

			buffer.WriteByte(str[i])
		}
	}

	if state == 0 {
		buffer.WriteString(gEnd)
	}

	buffer.WriteString("\n;\n")
	buffer.WriteString(fmt.Sprintf(": %s:print print ;\n", entry))

	return fc.Parse(buffer.String(), filename)
}

func compile_s(s *Stack[string], str []rune) {
	if len(str) > 9 {
		compile_g(s, str)
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

// compile a( ABC) to g( ABC) sv:fromS
func compile_a(s *Stack[string], str []rune) {
	compile_g(s, str)
	s.Push("sv:fromS")
}

// Loads a forth string onto the stack and terminates it with zero.
// The top element is the length of the string plus one.
// g( ABC) results in: 0 C B A 3
func compile_g(s *Stack[string], str []rune) {
	s.Push("0")

	for _, i := range reverse(str) {
		s.Push(fmt.Sprintf("%d", int(i)))
	}

	s.Push(fmt.Sprintf("%d", len(str)))
}

func handleForthString(s *Stack[string], str []rune) {
	switch str[0] {
	case '.':
		compile_s(s, str[3:len(str)-1])
	case 'a':
		compile_a(s, str[3:len(str)-1])
	case 'g':
		compile_g(s, str[3:len(str)-1])
	}
}

func IsFile(filename string) bool {
	info, err := os.Stat(filename)
	return !os.IsNotExist(err) && !info.IsDir()
}

func ListFiles(dir, ext string) ([]string, error) {
	files := make([]string, 0, 10)

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (fc *ForthCompiler) ReadFile(filename string) ([]byte, error) {
	if strings.Index(filename, "http://") == 0 ||
		strings.Index(filename, "https://") == 0 {
		resp, err := http.Get(filename)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}

	if !IsFile(filename) {
		found := false
		ext := filepath.Ext(filename)

		if ext == "" {
			ext = ".fs"
		}

		files, err := ListFiles(ConfigPath()+"lib/", ext)

		if err != nil {
			return nil, err
		}

		for _, path := range files {
			if strings.Contains(filepath.Base(path), filename) {
				filename = path
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("file \"%s\" not found", filename)
		}
	}

	return os.ReadFile(filename)
}

func (fc *ForthCompiler) ParseFile(filename string) error {
	data, err := fc.ReadFile(filename)

	if err != nil {
		return err
	}

	return fc.Parse(string(data), filename)
}

func (fc *ForthCompiler) ParseTemplateFile(entry, filename string) error {
	data, err := fc.ReadFile(filename)

	if err != nil {
		return err
	}

	return fc.ParseTemplate(entry, string(data), filename)
}

func (fc *ForthCompiler) handleMeta(meta string) error {
	cmd := strings.Split(meta, " ")

	switch cmd[0] {
	case "use":
		return fc.ParseFile(cmd[1])
	case "variable":
		if !fc.vars.Contains(cmd[1]) {
			fc.vars.Push(cmd[1])
		}
	case "template":
		return fc.ParseTemplateFile(cmd[1], cmd[2])
	default:
		return fmt.Errorf("unknown meta command \"%s\"", cmd[0])
	}

	return nil
}

func (fc *ForthCompiler) compileClass(def *Stack[string], filename string) error {
	if len(def.data) < 3 {
		return fmt.Errorf("a class must have at least one property")
	}

	clazz := def.data[0]

	if def.data[1] == "extends" {
		base := def.data[2]

		if _, ok := fc.defs[base+":sizeof"]; !ok {
			return fmt.Errorf("no base class \"%s\" found", base)
		}

		if err := fc.compileExtendedClass(clazz, base, filename); err != nil {
			return err
		}

		if len(def.data) > 3 {
			return fc.compileBasicClass(clazz, base, filename, def.data[3:])
		}

		return nil
	}

	return fc.compileBasicClass(clazz, "", filename, def.data[1:])
}

// class moo extends foo <1 a 1 b ...>
func (fc *ForthCompiler) compileExtendedClass(clazz, base, filename string) error {
	var builder strings.Builder
	substr := base + ":"

	for k := range fc.defs {
		if strings.Index(k, substr) == 0 {
			after, _ := strings.CutPrefix(k, substr) // ignore found, strings.index(k, substr) == 0
			builder.WriteString(fmt.Sprintf(": %s:%s %s ;\n", clazz, after, k))
		}
	}

	return fc.Parse(builder.String(), filename)
}

// class foo 1 a <1 b 5 c>
func (fc *ForthCompiler) compileBasicClass(clazz, base, filename string, values []string) error {
	var (
		builder strings.Builder
		offset  int64
		names   []string
		sizes   []int64
	)

	if def, ok := fc.defs[base+":sizeof"]; len(base) > 0 && ok {
		var err error

		if offset, err = strconv.ParseInt(def.data[0], 10, 64); err != nil {
			return err
		}
	}

	names = make([]string, 0, 10)
	sizes = make([]int64, 0, 10)

	for i := 0; i < len(values); i += 2 {
		name := values[i+1]
		names = append(names, name)
		size, err := strconv.ParseInt(values[i], 10, 64)
		sizes = append(sizes, size)

		if err != nil {
			return err
		}

		if size < 1 {
			return fmt.Errorf("struct member size must be greater than 0. Size was: %d at member: %s", size, name)
		}

		if offset > 0 {
			if offset == 1 {
				builder.WriteString(fmt.Sprintf(": %s:%s 1+ ;\n", clazz, name))
			} else {
				builder.WriteString(fmt.Sprintf(": %s:%s %d + ;\n", clazz, name, offset))
			}
		} else {
			builder.WriteString(fmt.Sprintf(": %s:%s ;\n", clazz, name))
		}

		offset += size
	}

	// clazz:init
	builder.WriteString(fmt.Sprintf(": %s:init ", clazz))
	for i, n := range names {
		if sizes[i] > 1 {
			builder.WriteString(fmt.Sprintf(" dup %d 0 rot %s:%s memset", sizes[i], clazz, n))
		} else {
			builder.WriteString(fmt.Sprintf(" dup 0 swap %s:%s !", clazz, n))
		}
	}
	if len(base) > 0 {
		builder.WriteString(fmt.Sprintf(" %s:init", base))
	}
	builder.WriteString(" ;\n")

	// basic methods
	builder.WriteString(fmt.Sprintf(": %s:sizeof %d ;\n", clazz, offset))
	builder.WriteString(fmt.Sprintf(": %s:allot %s:sizeof * allot ;\n", clazz, clazz))
	builder.WriteString(fmt.Sprintf(": %s:new 1 %s:allot %s:init ;\n", clazz, clazz, clazz))
	builder.WriteString(fmt.Sprintf(": %s:[] swap %s:sizeof * + ;\n", clazz, clazz))
	return fc.Parse(builder.String(), filename)
}

func (fc *ForthCompiler) compileLocals(iter *StackIter[string], result *Stack[string]) {
	localDefs := NewStack[string]()
	result.Push("LCTX")

	for iter.Next() {
		word := iter.Get()

		if word == "}" {
			break
		}

		if _, ok := fc.defs[word]; ok {
			PrintWarning(fmt.Sprintf("local %s shadows word with the same name", word))
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
		} else if word2 == "char" {
			iter.Next()
			word2 = iter.Get()
			if len(word2) > 1 {
				return fmt.Errorf("unable to get code point: \"%s\" is not a one-character", word2)
			}
			result.Push(fmt.Sprintf("L %d", int(word2[0])))
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
	// [-]?[0-9]+[\.][0-9]+?
	var (
		state  int
		result bool
	)

	for _, i := range s {
		switch state {
		case 0:
			if i == '0' {
				state = 3
			} else if i == '-' {
				state = 1
			} else if i >= '1' && i <= '9' {
				state = 2
			} else {
				return false
			}
		case 1:
			if i == '0' {
				state = 3
			} else if i >= '1' && i <= '9' {
				state = 2
			} else {
				return false
			}
		case 2:
			if i == '.' {
				state = 4
				result = true
			} else if i < '0' || i > '9' {
				return false
			}
		case 3:
			if i == '.' {
				state = 4
				result = true
			} else {
				return false
			}
		case 4:
			if i < '0' || i > '9' {
				return false
			}
		}
	}

	return result
}

func isNumeric(s string) bool {
	// [-]?[0-9]+
	var (
		state  int
		result bool
	)

	for _, i := range s {
		switch state {
		case 0:
			if i == '0' {
				state = 3
				result = true
			} else if i == '-' {
				state = 1
			} else if i >= '1' && i <= '9' {
				state = 2
				result = true
			} else {
				return false
			}
		case 1:
			if i >= '1' && i <= '9' {
				state = 2
				result = true
			} else {
				return false
			}
		case 2:
			if i < '0' || i > '9' {
				return false
			}
		case 3:
			return false
		}
	}

	return result
}

func isString(s string) bool {
	var (
		state int
	)

	for index, i := range s {
		switch state {
		case 0:
			switch i {
			case '.', 'a', 'g':
				if index+1 == len(s) {
					break
				}
				state = 1
			default:
				return false
			}
		case 1:
			switch i {
			case '"', '(':
				state = 2
			default:
				return false
			}
		case 2:
			// consume space
			switch i {
			case ' ':
				state = 3
			default:
				return false
			}
		case 3:
			if index == len(s)-1 && (i == ')' || i == '"') {
				return true
			}
			// consume string
		}
	}

	return false
}

func (fc *ForthCompiler) compileWord(word string, result *Stack[string]) error {
	if isString(word) {
		tmp := NewStack[string]()
		handleForthString(tmp, []rune(word))
		for value := range tmp.All() {
			fc.compileWord(value, result)
		}
	} else if isNumeric(word) {
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
