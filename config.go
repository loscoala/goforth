package goforth

import (
	_ "embed"
)

// The core words dictionary
//
//go:embed core.fs
var Core string

// Colored output
var Colored bool

// The prompt in StartREPL
var Repl = "forth> "

// Show byte code in StartREPL
var ShowByteCode bool

// Show execution time in vm.Run
var ShowExecutionTime bool
