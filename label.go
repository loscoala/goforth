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
