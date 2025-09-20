package goforth

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var baseSyntax = [...]string{
	"begin", "while", "repeat", "do", "?do", "loop", "+loop", "-loop", "if", "then",
	"else", "{", "}", "[", "]", "until", "again", "leave", "to", "done", ":", ";",
	"case", "of", "?of", "endof", "endcase", "variable", "char", "class", "extends",
	"inline",
}

var (
	Magenta = color.New(color.FgHiMagenta).SprintFunc()
	Cyan    = color.New(color.FgHiCyan).SprintFunc()
	Green   = color.New(color.FgHiGreen).SprintFunc()
	Blue    = color.New(color.FgHiBlue).SprintFunc()
	Yellow  = color.New(color.FgHiYellow).SprintFunc()
	Red     = color.New(color.FgHiRed).SprintFunc()
)

func isBaseSytax(word string) bool {
	for _, w := range baseSyntax {
		if w == word {
			return true
		}
	}

	return false
}

func getWordColored(fc *ForthCompiler, word string) string {
	if _, ok := fc.data[word]; ok {
		return Magenta(word)
	} else if _, ok := fc.defs[word]; ok {
		return Cyan(word)
	} else if _, ok := fc.inlines[word]; ok {
		return Red(word)
	} else if fc.vars.Contains(word) {
		return Red(word)
	} else if isBaseSytax(word) {
		return Green(word)
	} else if isFloat(word) || isNumeric(word) {
		return Blue(word)
	} else if isString(word) {
		return Cyan(word)
	}
	return Yellow(word)
}

func printVariableColored(fc *ForthCompiler, word string) {
	fmt.Printf("%s %s\n", getWordColored(fc, "variable"), getWordColored(fc, word))
}

func printVariable(word string) {
	fmt.Printf("variable %s\n", word)
}

func printWordColored(fc *ForthCompiler, word string, s *Stack[string]) {
	fmt.Printf("%s %s ", getWordColored(fc, ":"), getWordColored(fc, word))

	for iter := s.Iter(); iter.Next(); {
		fmt.Printf("%s ", getWordColored(fc, iter.Get()))
	}

	fmt.Printf("%s\n", getWordColored(fc, ";"))
}

func printWord(word string, s *Stack[string]) {
	fmt.Printf(": %s ", word)

	for iter := s.Iter(); iter.Next(); {
		fmt.Printf("%s ", iter.Get())
	}

	fmt.Println(";")
}

func PrintError(err error) {
	if Colored {
		fmt.Printf("%s: %s\n", Red("[Error]"), err)
	} else {
		fmt.Printf("[Error]: %s\n", err)
	}
}

func PrintWarning(warning string) {
	if Colored {
		fmt.Printf("%s: %s\n", Yellow("[Warning]"), warning)
	} else {
		fmt.Printf("[Warning]: %s\n", warning)
	}
}

func (fc *ForthCompiler) findDefinitions(word string) *Stack[string] {
	result := NewStack[string]()

	f := func(data map[string]*Stack[string], word string, result *Stack[string]) {
		for k, v := range data {
			for value := range v.Values() {
				if value == word {
					if !result.Contains(k) {
						result.Push(k)
					}
				}
			}

			if strings.Contains(k, word) {
				result.Push(k)
			}
		}
	}

	f(fc.defs, word, result)
	f(fc.inlines, word, result)

	sort.Strings(result.data)
	return result
}

func (fc *ForthCompiler) printDefinition(word string) {
	if fc.vars.Contains(word) {
		if Colored {
			printVariableColored(fc, word)
		} else {
			printVariable(word)
		}
		return
	}

	s, ok := fc.defs[word]

	if !ok {
		s, ok = fc.inlines[word]
	}

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
					fmt.Printf("Unknown word %s\n", Red(word))
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
	var wg sync.WaitGroup
	keys := make([]string, 0, len(fc.defs))
	mkeys := make([]string, 0, len(fc.inlines))

	f_keys := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for k := range fc.defs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}

	f_mkeys := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for k := range fc.inlines {
			mkeys = append(mkeys, k)
		}
		sort.Strings(mkeys)
	}

	wg.Add(2)
	go f_keys(&wg)
	go f_mkeys(&wg)

	wg.Wait()

	if Colored {
		for val := range fc.vars.Values() {
			printVariableColored(fc, val)
		}

		for _, k := range keys {
			printWordColored(fc, k, fc.defs[k])
		}

		for _, k := range mkeys {
			printWordColored(fc, k, fc.inlines[k])
		}
	} else {
		for val := range fc.vars.Values() {
			printVariable(val)
		}

		for _, k := range keys {
			printWord(k, fc.defs[k])
		}

		for _, k := range mkeys {
			printWord(k, fc.inlines[k])
		}
	}
}

func (fc *ForthCompiler) printByteCode() {
	if Colored {
		for cmd := range strings.SplitSeq(fc.ByteCode(), ";") {
			if cmd == "" {
				continue
			}
			fmt.Printf("%s;", Yellow(cmd))
			if cmd == "END" || strings.Index(cmd, "GDEF ") == 0 {
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

func (fc *ForthCompiler) handleStdin() {
	scanner := bufio.NewScanner(os.Stdin)
	data := make([]byte, 0, 1000)

	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}

	if err := scanner.Err(); err != nil {
		PrintError(err)
		return
	}

	if err := fc.Run(string(data)); err != nil {
		PrintError(err)
		return
	}
}

func (fc *ForthCompiler) initCompleter() readline.AutoCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, 100)

	for k := range fc.defs {
		items = append(items, readline.PcItem(k))
	}

	c := readline.NewPrefixCompleter(items...)

	return c
}

func (fc *ForthCompiler) initReadline() *readline.Instance {
	instance, err := readline.NewEx(&readline.Config{
		Prompt:          Repl,
		InterruptPrompt: "^C",
		EOFPrompt:       "type 'exit' to quit",
		AutoComplete:    fc.initCompleter(),
	})

	if err != nil {
		panic(err)
	}

	return instance
}

func (fc *ForthCompiler) handleREPL() {
	var text string

	line := fc.initReadline()
	defer line.Close()
	line.CaptureExitSignal()

	for {
		// encapsulate err
		{
			var err error
			text, err = line.Readline()

			if err == readline.ErrInterrupt {
				if len(text) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}
		}

		if isWhiteSpace(text) {
			continue
		}

		if text == "exit" {
			break
		}

		if text[0] == ':' {
			// just parse
			if err := fc.Parse(text, "main"); err != nil {
				PrintError(err)
			}
			line.Config.AutoComplete = fc.initCompleter()
			continue
		}

		if text[0] == '%' && len(text) > 1 {
			fc.printDefinition(text[2:])
			continue
		} else if text[0] == '%' && len(text) == 1 {
			fc.printAllDefinitions()
			continue
		} else if strings.Index(text, "find ") == 0 {
			defs := fc.findDefinitions(text[5:])
			for value := range defs.Values() {
				fc.printDefinition(value)
			}
			continue
		} else if strings.Index(text, "use ") == 0 ||
			strings.Index(text, "template ") == 0 {
			if err := fc.handleMeta(text); err != nil {
				PrintError(err)
			}
			line.Config.AutoComplete = fc.initCompleter()
			continue
		} else if strings.Index(text, "reset") == 0 {
			clear(fc.defs)
			clear(fc.inlines)
			fc.ParseFile("core")
			continue
		} else if strings.Index(text, "variable ") == 0 {
			if err := fc.handleMeta(text); err != nil {
				PrintError(err)
			}
			continue
		} else if strings.Index(text, "debug ") == 0 {
			if err := fc.Parse(": main\n"+text[6:]+"\n;", "main"); err != nil {
				PrintError(err)
				continue
			}

			if err := fc.Preprocess(); err != nil {
				PrintError(err)
				continue
			}

			if err := fc.Compile(); err != nil {
				PrintError(err)
				continue
			}

			fc.printByteCode()
			fmt.Println("")
			fc.printDebug()
			continue
		} else if strings.Index(text, "compile ") == 0 {
			if err := fc.Parse(": main\n"+text[8:]+"\n;", "main"); err != nil {
				PrintError(err)
				continue
			}

			if err := fc.Preprocess(); err != nil {
				PrintError(err)
				continue
			}

			if err := fc.Compile(); err != nil {
				PrintError(err)
				continue
			}

			if ShowByteCode {
				fc.printByteCode()
				fmt.Println("")
			}

			if err := fc.CompileToC(); err != nil {
				PrintError(err)
			}

			continue
		} else if strings.Index(text, "pp") == 0 {
			if err := fc.Preprocess(); err != nil {
				PrintError(err)
				continue
			}

			continue
		}

		if err := fc.Parse(": main\n"+text+"\n;", "main"); err != nil {
			PrintError(err)
			continue
		}

		if err := fc.Preprocess(); err != nil {
			PrintError(err)
			continue
		}

		if err := fc.Compile(); err != nil {
			PrintError(err)
			continue
		}

		if ShowByteCode {
			fc.printByteCode()
		}

		// skip empty code
		if fc.defs["main"].Len() == 0 {
			continue
		}

		fc.Fvm.Run(fc.ByteCode())

		if fc.Fvm.ExitStatus != 0 {
			PrintError(fmt.Errorf("exit status: %d", fc.Fvm.ExitStatus))
		} else {
			fmt.Println("")
		}
	}
}

func randomStringBytes(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = byte(65 + rand.Intn(26))
	}

	return string(b)
}

func initNameCache() func(name string) string {
	cache := make(map[string]string)

	return func(name string) string {
		if name == "" {
			// return all names
			var result strings.Builder
			for k, v := range cache {
				result.WriteString(fmt.Sprintf("static void %s(void); // %s\n", v, k))
			}
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			return result.String()
		}

		if ret, ok := cache[name]; ok {
			return ret
		}

		r := "fvm_" + randomStringBytes(6)

		cache[name] = r
		return r
	}
}

func (fc *ForthCompiler) initGlobalNameCache() func(name string) string {
	cache := make(map[string]string)

	return func(name string) string {
		if name == "" {
			// return all names
			var result strings.Builder
			for k, v := range cache {
				result.WriteString(fmt.Sprintf("static cell_t %s = { .value = %d }; // %s\n", v, 0, k))
			}
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			return result.String()
		}

		if ret, ok := cache[name]; ok {
			return ret
		}

		r := "g_" + randomStringBytes(6)

		cache[name] = r
		return r
	}
}

func initVarNameCache() func(name string) string {
	cache := make(map[string]string)

	return func(name string) string {
		if ret, ok := cache[name]; ok {
			return ret
		}

		r := "v_" + randomStringBytes(6)

		cache[name] = r
		return r
	}
}

// returns spaces. indent can only be a multible of 2 and between 2 and 20.
func initSpaceCache() func(indent int) string {
	spaces := [...]string{
		"  ",
		"    ",
		"      ",
		"        ",
		"          ",
		"            ",
		"              ",
		"                ",
		"                  ",
		"                    ",
	}

	return func(indent int) string {
		return spaces[((indent/2)-1)%10]
	}
}

func (fc *ForthCompiler) CompileToC() error {
	var result strings.Builder
	funcs := initNameCache()
	locals := initVarNameCache()
	globals := fc.initGlobalNameCache()
	spaces := initSpaceCache()
	indent := 2

	{
		m := fc.defs["main"]
		result.WriteString("// compiled from:\n//")
		for word := range m.Values() {
			result.WriteString(" ")
			result.WriteString(word)
		}
		result.WriteString("\n\n")
	}

	for _, cmd := range strings.Split(fc.ByteCode(), ";") {
		if cmd == "" {
			continue
		}

		scmd := strings.Split(cmd, " ")

		if len(scmd) == 2 && scmd[0][0] == '#' {
			// NOP
			result.WriteString(fmt.Sprintf("l%s:\n%s;\n", scmd[0][1:], spaces(indent)))
		} else {
			switch scmd[0] {
			case "GDEF":
				globals(scmd[1])
			case "GSET":
				result.WriteString(fmt.Sprintf("%s%s = fvm_pop(); // %s\n", spaces(indent), globals(scmd[1]), scmd[1]))
			case "GBL":
				result.WriteString(fmt.Sprintf("%sfvm_push(%s); // %s\n", spaces(indent), globals(scmd[1]), scmd[1]))
			case "JMP":
				result.WriteString(fmt.Sprintf("%sgoto l%s;\n", spaces(indent), scmd[1][1:]))
			case "JIN":
				result.WriteString(fmt.Sprintf("%sif (fvm_jin()) goto l%s;\n", spaces(indent), scmd[1][1:]))
			case "L":
				result.WriteString(fmt.Sprintf("%sfvm_push(fvm_cell(%s));\n", spaces(indent), scmd[1]))
			case "LF":
				result.WriteString(fmt.Sprintf("%sfvm_push(fvm_cell_d(%s));\n", spaces(indent), scmd[1]))
			case "LCTX":
				result.WriteString(fmt.Sprintf("%s{\n", spaces(indent)))
				indent += 2
			case "LCLR":
				indent -= 2
				result.WriteString(fmt.Sprintf("%s}\n", spaces(indent)))
			case "LDEF":
				result.WriteString(fmt.Sprintf("%scell_t %s = fvm_pop(); // %s\n", spaces(indent), locals(scmd[1]), scmd[1]))
			case "LCL":
				result.WriteString(fmt.Sprintf("%sfvm_push(%s); // %s\n", spaces(indent), locals(scmd[1]), scmd[1]))
			case "LSET":
				result.WriteString(fmt.Sprintf("%s%s = fvm_pop(); // %s\n", spaces(indent), locals(scmd[1]), scmd[1]))
			case "SUB":
				result.WriteString(fmt.Sprintf("static void %s(void) { // %s\n", funcs(scmd[1]), scmd[1]))
			case "END":
				result.WriteString("}\n\n")
			case "MAIN":
				result.WriteString("int main(int argc, char** argv) {\n  fvm_argc = (int64_t)argc;\n  fvm_argv = argv;\n")
				if ShowExecutionTime {
					result.WriteString("  fvm_time();\n")
				}
			case "CALL":
				result.WriteString(fmt.Sprintf("%s%s(); // %s\n", spaces(indent), funcs(scmd[1]), scmd[1]))
			case "REF":
				result.WriteString(fmt.Sprintf("%sfvm_ref(&%s); // %s\n", spaces(indent), funcs(scmd[1]), scmd[1]))
			default:
				result.WriteString(fmt.Sprintf("%sfvm_%s();\n", spaces(indent), strings.ToLower(scmd[0])))
			}
		}
	}

	result.WriteString("  return 0;\n}\n")

	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		if err := os.Mkdir(ConfigPath(), 0750); err != nil {
			return err
		} else {
			if err := os.WriteFile(ConfigPath()+"vm.c", CVM, 0644); err != nil {
				return err
			}
		}
	}

	if err := os.WriteFile(ConfigPath()+CCodeName, []byte("#include \"vm.c\"\n\n"+funcs("")+globals("")+result.String()), 0644); err != nil {
		return err
	}

	result.Reset()

	if CAutoCompile {
		if err := fc.compileToBinary(); err != nil {
			return err
		}

		if CAutoExecute {
			if err := fc.runBinary(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (fc *ForthCompiler) compileToBinary() error {
	cmd := exec.Command(CCompiler, "-o", CBinaryName, CCodeName, COptimization)
	cmd.Dir = ConfigPath()
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (fc *ForthCompiler) runBinary() error {
	cmd := exec.Command(ConfigPath() + CBinaryName)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func (fc *ForthCompiler) printDebug() {
	var (
		out strings.Builder
		ok  bool
		err error
	)

	oldOut := fc.Fvm.Out
	fc.Fvm.Out = &out

	fc.Fvm.PrepareRun(fc.ByteCode())
	fmt.Printf("    %-15s | %-25s | %-25s | %s\n", "LINE, OP", "STACK", "RSTACK", "OUTPUT")

	for !ok {
		fmt.Printf("%3d ", fc.Fvm.CodeData.ProgPtr)

		if ok, err = fc.Fvm.RunStep(); err != nil {
			PrintError(err)
			break
		} else {
			cmd := fc.Fvm.CodeData.Command.String()

			output := out.String()
			if output == "\n" || output == "\r\n" {
				output = "\\n"
			}

			fmt.Printf("%-15s | %-25s | %-25s | %s\n",
				cmd,
				strings.Trim(fmt.Sprintf("%v", fc.Fvm.Stack), "[]"),
				strings.Trim(fmt.Sprintf("%v", fc.Fvm.Rstack), "[]"),
				output)
			out.Reset()
		}
	}

	if fc.Fvm.ExitStatus != 0 {
		fmt.Printf("Exit status: %d\n", fc.Fvm.ExitStatus)
	}

	fc.Fvm.Out = oldOut
}

func (fc *ForthCompiler) StartREPL() {
	stat, err := os.Stdin.Stat()

	if err != nil {
		PrintError(err)
		return
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		fc.handleStdin()
	} else {
		fc.handleREPL()
	}
}

func (fc *ForthCompiler) RunFile(str string) error {
	if err := fc.ParseFile(str); err != nil {
		return err
	}

	if err := fc.Preprocess(); err != nil {
		return err
	}

	if err := fc.Compile(); err != nil {
		return err
	}

	fc.Fvm.Run(fc.ByteCode())

	return nil
}

func (fc *ForthCompiler) CompileFile(str string) error {
	if err := fc.ParseFile(str); err != nil {
		return err
	}

	if err := fc.Preprocess(); err != nil {
		return err
	}

	if err := fc.Compile(); err != nil {
		return err
	}

	return fc.CompileToC()
}

func (fc *ForthCompiler) CompileScript(script string) error {
	if err := fc.Parse(script, "script"); err != nil {
		return err
	}

	if err := fc.Preprocess(); err != nil {
		return err
	}

	if err := fc.Compile(); err != nil {
		return err
	}

	return fc.CompileToC()
}

func (fc *ForthCompiler) Run(prog string) error {
	if err := fc.Parse(prog, "script"); err != nil {
		return err
	}

	if err := fc.Preprocess(); err != nil {
		return err
	}

	if err := fc.Compile(); err != nil {
		return err
	}

	fc.Fvm.Run(fc.ByteCode())

	return nil
}
