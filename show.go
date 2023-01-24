package goforth

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var baseSyntax = [...]string{
	"begin", "while", "repeat", "do", "?do", "loop", "+loop", "-loop", "if", "then",
	"else", "{", "}", "[", "]", "until", "again", "leave", "to", "done", ":", ";",
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
	} else if isBaseSytax(word) {
		return Green(word)
	} else if isFloat(word) || isNumeric(word) {
		return Blue(word)
	} else {
		return Yellow(word)
	}
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
		fmt.Printf("%s: %s\n", Red("[Error]"), err)
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
			fmt.Printf("%s;", Yellow(cmd))
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
			if err := fc.Parse(text); err != nil {
				PrintError(err)
			}
			line.Config.AutoComplete = fc.initCompleter()
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
		} else if strings.Index(text, "use ") == 0 {
			if err := fc.ParseFile(text[4:]); err != nil {
				PrintError(err)
			}
			line.Config.AutoComplete = fc.initCompleter()
			continue
		} else if strings.Index(text, "debug ") == 0 {
			if err := fc.Parse(": main\n" + text[6:] + "\n;"); err != nil {
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
			if err := fc.Parse(": main\n" + text[8:] + "\n;"); err != nil {
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
			fc.compileToC()
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

		if ShowByteCode {
			fc.printByteCode()
		}

		// skip empty code
		if fc.defs["main"].Len() == 0 {
			continue
		}

		fc.Fvm.Run(fc.ByteCode())
		fmt.Println("")
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

func (fc *ForthCompiler) compileToC() {
	var result strings.Builder
	funcs := initNameCache()

	for _, cmd := range strings.Split(fc.ByteCode(), ";") {
		if cmd == "" {
			continue
		}

		scmd := strings.Split(cmd, " ")

		if len(scmd) == 2 && scmd[0][0] == '#' {
			// NOP
			result.WriteString(fmt.Sprintf("l%s:\n", scmd[0][1:]))
		} else {
			switch scmd[0] {
			case "JMP":
				result.WriteString(fmt.Sprintf("  goto l%s;\n", scmd[1][1:]))
			case "JIN":
				result.WriteString(fmt.Sprintf("  if (fvm_jin()) goto l%s;\n", scmd[1][1:]))
			case "L":
				result.WriteString(fmt.Sprintf("  fvm_push(%s);\n", scmd[1]))
			case "LF":
				// TODO: currently it only works with valied floats in Forth like 1.0 and not 1.
				result.WriteString(fmt.Sprintf("  fvm_fpush(%s);\n", scmd[1]))
			case "STP":
				if ShowExecutionTime {
					result.WriteString("\n  end = clock();\n  time_spend = (double)(end - begin);\n  printf(\"\\ntime: %fs\\n\", time_spend / CLOCKS_PER_SEC);\n")
				}
				result.WriteString("  fvm_stp();\n}\n")
			case "SUB":
				result.WriteString(fmt.Sprintf("static void %s(void) { // %s\n", funcs(scmd[1]), scmd[1]))
			case "END":
				result.WriteString("  return;\n}\n\n")
			case "MAIN":
				result.WriteString("int main(int argc, char** argv) {\n")
				if ShowExecutionTime {
					result.WriteString("  clock_t begin, end;\n  double time_spend;\n\n  begin = clock();\n\n")
				}
			case "CALL":
				result.WriteString(fmt.Sprintf("  %s(); // %s\n", funcs(scmd[1]), scmd[1]))
			case "REF":
				result.WriteString(fmt.Sprintf("  fvm_ref(&%s); // %s\n", funcs(scmd[1]), scmd[1]))
			default:
				result.WriteString(fmt.Sprintf("  fvm_%s();\n", strings.ToLower(scmd[0])))
			}
		}
	}

	timeheader := ""
	if ShowExecutionTime {
		timeheader = "#include <time.h>\n\n"
	}

	os.WriteFile("lib/"+CCodeName, []byte("#include \"vm.c\"\n\n"+timeheader+funcs("")+result.String()), 0644)

	if CAutoCompile {
		cmd := exec.Command(CCompiler, "-o", CBinaryName, CCodeName, COptimization)
		cmd.Dir = "lib/"

		if err := cmd.Run(); err != nil {
			PrintError(err)
		} else if CAutoExecute {
			cmd := exec.Command("lib/" + CBinaryName)

			if out, err := cmd.Output(); err != nil {
				PrintError(err)
			} else {
				fmt.Println(string(out))
			}
		}
	}
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
