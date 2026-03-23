package goforth

import (
	"embed"
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

// The name of the C code file
var CCodeName = "main.c"

// The name of the binary
var CBinaryName = "main"

// If set use the current path as output directory
var CCurrentDir bool

// The stdlib in goforth and runtime in C
//
//go:embed stdlib/*.fs lib/vm.c
var Stdlib embed.FS

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

	// Create goforth dir in user config path
	if _, err := os.Stat(cachedConfigPath); os.IsNotExist(err) {
		if err2 := os.Mkdir(cachedConfigPath, 0750); err2 != nil {
			log.Fatal(err2)
		}
	}

	return cachedConfigPath
}
