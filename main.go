package main

import (
	_ "embed"
	"flag"
)

var (
	colored bool
	fname   string
	script  string
	//go:embed core.fs
	corefs string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "compile file and execute")
	flag.BoolVar(&colored, "color", true, "Use colors")
	flag.StringVar(&script, "script", "", "program passed in as string")

	flag.Parse()
}

func main() {
	initFlags()

	fc := NewForthCompiler()

	// fc.ParseFile(path.Join(path.Dir(os.Args[0]), "core.fs"))
	if err := fc.Parse(corefs); err != nil {
		printError(err)
	}

	if len(script) > 0 {
		if err := fc.Run(script); err != nil {
			printError(err)
		}
	} else if len(fname) > 0 {
		if err := fc.RunFile(fname); err != nil {
			printError(err)
		}
	} else {
		fc.StartREPL()
	}
}
