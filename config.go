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

// The name of the C compiler
var CCompiler = "cc"

// The optimization flag of the C compiler
var COptimization = "-O2"

// Compile automatically after C code generation
var CAutoCompile = true

// Automatically execute the binary after compiling
var CAutoExecute = true

// The name of the C code file
var CCodeName = "main.c"

// The name of the binary
var CBinaryName = "main"
