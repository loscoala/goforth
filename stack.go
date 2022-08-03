package main

import (
	"log"
)

type Stack struct {
	data []string
}

func (s *Stack) Len() int {
	return len(s.data)
}

func (s *Stack) IsEmpty() bool {
	return s.Len() == 0
}

func (s *Stack) Push(str string) {
	s.data = append(s.data, str)
}

func (s *Stack) Pop() (string, bool) {
	if s.IsEmpty() {
		return "", false
	} else {
		index := s.Len() - 1
		element := s.data[index]
		s.data = s.data[:index]
		return element, true
	}
}

func (s *Stack) ExPop() string {
	value, ok := s.Pop()
	if !ok {
		log.Fatal("Error: Pop() from empty Stack")
	}
	return value
}

func (s *Stack) Fetch() (string, bool) {
	if s.IsEmpty() {
		return "", false
	}

	index := s.Len() - 1
	element := s.data[index]
	return element, true
}

func (s *Stack) ExFetch() string {
	value, ok := s.Fetch()
	if !ok {
		log.Fatal("Error: Fetch() from empty Stack")
	}
	return value
}

func (s *Stack) Reverse() *Stack {
	var result Stack

	for s.Len() > 0 {
		if value, ok := s.Pop(); ok {
			result.Push(value)
		}
	}

	return &result
}

/*
func (s *Stack) Append(stk *Stack) {
	nstk := stk.Reverse()
	for nstk.Len() > 0 {
		if value, ok := nstk.Pop(); ok {
			s.Push(value)
		}
	}
}
*/

type StackIter struct {
	stack *Stack
	len   int
	index int
}

func (s *Stack) Iter() *StackIter {
	return &StackIter{stack: s, len: s.Len(), index: -1}
}

func (s *StackIter) Next() bool {
	if s.index < s.len-1 {
		s.index++
		return true
	}

	return false
}

func (s *StackIter) Get() string {
	return s.stack.data[s.index]
}

func (s *Stack) Contains(str string) bool {
	for _, i := range s.data {
		if i == str {
			return true
		}
	}

	return false
}

func (s *Stack) GetIndex(str string) int {
	for pos, i := range s.data {
		if i == str {
			return pos
		}
	}

	return -1
}

// -------------------- SliceStack ----------------------------

type SliceStack []*Stack

func (ss *SliceStack) Contains(str string) bool {
	for _, i := range *ss {
		if i.Contains(str) {
			return true
		}
	}

	return false
}

func (ss *SliceStack) Len() int {
	return len(*ss)
}

func (ss *SliceStack) IsEmpty() bool {
	return ss.Len() == 0
}

func (ss *SliceStack) Push(stk *Stack) {
	*ss = append(*ss, stk)
}

func (ss *SliceStack) Pop() (*Stack, bool) {
	if ss.IsEmpty() {
		return nil, false
	} else {
		index := ss.Len() - 1
		element := (*ss)[index]
		*ss = (*ss)[:index]
		return element, true
	}
}

func (ss *SliceStack) ExPop() *Stack {
	value, ok := ss.Pop()
	if !ok {
		log.Fatal("Error: Pop() from empty SliceStack")
	}
	return value
}
