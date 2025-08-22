package goforth

import (
	_ "embed"
	"log"
	"os"
)

// Colored output
var Colored bool

// The prompt in StartREPL
var Repl = Magenta("forth> ")

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

// The vm in C
//
//go:embed lib/vm.c
var CVM []byte

var cachedConfigPath string

func ConfigPath() string {
	if cachedConfigPath != "" {
		return cachedConfigPath
	}

	dir, err := os.UserConfigDir()

	if err != nil {
		log.Fatal(err)
	}

	cachedConfigPath = dir + "/goforth/"
	return cachedConfigPath
}
