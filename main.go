package main

import (
	_ "embed"
	"flag"
)

var (
	colored bool
	fname   string
	//go:embed core.fs
	corefs string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "compile file and execute")
	flag.BoolVar(&colored, "color", true, "Use colors")

	flag.Parse()
}

func main() {
	initFlags()

	fc := NewForthCompiler()

	// fc.ParseFile(path.Join(path.Dir(os.Args[0]), "core.fs"))
	if err := fc.Parse(corefs); err != nil {
		printError(err)
	}

	if len(fname) == 0 {
		fc.StartREPL()
	} else {
		if err := fc.RunFile(fname); err != nil {
			printError(err)
		}
	}
}
