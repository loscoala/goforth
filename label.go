package main

import (
	"fmt"
)

type Label struct {
	label int
}

func (l *Label) CreateNewLabel() string {
	lbl := fmt.Sprintf("%d", l.label)
	l.label++
	return lbl
}

func (l *Label) CreateNewWord() string {
	lbl := fmt.Sprintf("b%d", l.label)
	l.label++
	return lbl
}
