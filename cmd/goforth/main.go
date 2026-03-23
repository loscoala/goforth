package main

import (
	"flag"

	"github.com/loscoala/goforth"
)

var (
	fname   string
	script  string
	compile bool
	outfile string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "Program passed in as a file")
	flag.BoolVar(&goforth.Colored, "color", true, "Use colors")
	flag.StringVar(&script, "script", "", "Program passed in as string")
	flag.BoolVar(&compile, "compile", false, "Compile to C")
	flag.StringVar(&outfile, "o", goforth.CBinaryName, "The name of the generated binary file (-compile flag is required)")

	flag.Parse()
}

func main() {
	initFlags()

	fc := goforth.NewForthCompiler()

	// custom sys func
	//fc.Fvm.Sysfunc = func(fvm *goforth.ForthVM, syscall int64) {
	//	switch syscall {
	//	case 999:
	//		fmt.Println("This is a custom call")
	//	default:
	//		fmt.Println("Not implemented")
	//	}
	//}

	// load the core words
	if err := fc.ParseFile("core"); err != nil {
		goforth.PrintError(err)
	}

	if outfile != goforth.CBinaryName {
		goforth.CBinaryName = outfile
		goforth.CCodeName = outfile + ".c"
		goforth.CCurrentDir = true
	}

	if len(script) > 0 {
		if compile {
			if err := fc.CompileScript(script); err != nil {
				goforth.PrintError(err)
			}
		} else {
			if err := fc.Run(script); err != nil {
				goforth.PrintError(err)
			}
		}
	} else if len(fname) > 0 {
		if compile {
			if err := fc.CompileFile(fname); err != nil {
				goforth.PrintError(err)
			}
		} else {
			if err := fc.RunFile(fname); err != nil {
				goforth.PrintError(err)
			}
		}
	} else {
		fc.StartREPL()
	}
}
