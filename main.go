package main

import (
	"flag"
)

var (
	colored bool
	fname   string
)

func initFlags() {
	flag.StringVar(&fname, "file", "", "compile file and execute")
	flag.BoolVar(&colored, "c", true, "Use colors")

	flag.Parse()
}

func main() {
	initFlags()

	// -----------------------

	fc := NewForthCompiler()

	// fc.Parse("\\ comment here \n: add2 2 + ;")
	fc.ParseFile("core.fs")

	if len(fname) == 0 {
		fc.StartREPL()
	} else {
		fc.RunFile(fname)
	}
}
