package main

import (
	"flag"
	"os"
	"path"
)

var (
	colored bool
	fname   string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "compile file and execute")
	flag.BoolVar(&colored, "color", true, "Use colors")

	flag.Parse()
}

func main() {
	initFlags()

	// -----------------------

	fc := NewForthCompiler()

	// fc.Parse("\\ comment here \n: add2 2 + ;")
	fc.ParseFile(path.Dir(os.Args[0]) + "/core.fs")

	if len(fname) == 0 {
		fc.StartREPL()
	} else {
		fc.RunFile(fname)
	}
}
