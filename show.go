package main

import (
	"fmt"
)

const (
	HEADER    = "\033[95m"
	OKBLUE    = "\033[94m"
	OKCYAN    = "\033[96m"
	OKGREEN   = "\033[92m"
	WARNING   = "\033[93m"
	FAIL      = "\033[91m"
	ENDC      = "\033[0m"
	BOLD      = "\033[91m"
	UNDERLINE = "\033[4m"
)

var baseSyntax = []string{
	"begin", "while", "repeat", "do", "?do", "loop", "+loop", "if", "then", "else", "{", "}", "until", "again", "leave", "to",
}

func isBaseSytax(word string) bool {
	for _, w := range baseSyntax {
		if w == word {
			return true
		}
	}

	return false
}

func getWordColored(fc *ForthCompiler, word string) string {
	var color string

	if _, ok := fc.data[word]; ok {
		color = HEADER
	} else if _, ok := fc.defs[word]; ok {
		color = OKCYAN
	} else if isBaseSytax(word) {
		color = OKGREEN
	} else if isNumeric(word) {
		color = OKBLUE
	} else {
		color = WARNING
	}

	return fmt.Sprintf("%s%s%s", color, word, ENDC)
}

func printWordColored(fc *ForthCompiler, word string, s *Stack) {
	fmt.Printf("%s %s ", getWordColored(fc, ":"), getWordColored(fc, word))

	for iter := s.Iter(); iter.Next(); {
		fmt.Printf("%s ", getWordColored(fc, iter.Get()))
	}

	fmt.Printf("%s\n", getWordColored(fc, ";"))
}

func (fc *ForthCompiler) printResult(s *Stack) {
	for iter := s.Iter(); iter.Next(); {
		if colored {
			fmt.Printf("%s%s%s;", WARNING, iter.Get(), ENDC)
		} else {
			fmt.Printf("%s;", iter.Get())
		}
		fc.output.WriteString(iter.Get() + ";")
	}

	fmt.Println("")
}

func printWord(word string, s *Stack) {
	fmt.Printf(": %s ", word)

	for iter := s.Iter(); iter.Next(); {
		fmt.Printf("%s ", iter.Get())
	}

	fmt.Println(";")
}

func printDefinition(fc *ForthCompiler, word string) {
	s := fc.defs[word]

	if colored {
		printWordColored(fc, word, s)
	} else {
		printWord(word, s)
	}
}

func printAllDefinitions(fc *ForthCompiler) {
	if colored {
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

func (fc *ForthCompiler) printSubs() {
	if len(fc.funcs) > 0 {
		for _, v := range fc.funcs {
			fc.printResult(v)
		}
		fc.funcs = make(map[string]*Stack)
		fmt.Println("")
	}
	if colored {
		fmt.Printf("%sMAIN%s;", WARNING, ENDC)
	} else {
		fmt.Printf("MAIN;")
	}
	fc.output.WriteString("MAIN;")
}
