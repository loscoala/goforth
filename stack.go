package goforth

import (
	"iter"
	"log"
	"slices"
)

type Stack[T comparable] struct {
	data []T
}

func NewStack[T comparable]() *Stack[T] {
	stack := new(Stack[T])
	stack.data = make([]T, 0, 100)
	return stack
}

func (s *Stack[T]) Len() int {
	return len(s.data)
}

func (s *Stack[T]) IsEmpty() bool {
	return s.Len() == 0
}

func (s *Stack[T]) Push(val T) {
	s.data = append(s.data, val)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T

	if s.IsEmpty() {
		return zero, false
	}

	index := s.Len() - 1
	element := s.data[index]
	s.data[index] = zero
	s.data = s.data[:index]
	return element, true
}

func (s *Stack[T]) ExPop() T {
	value, ok := s.Pop()
	if !ok {
		log.Fatal("Error: Pop() from empty Stack")
	}
	return value
}

func (s *Stack[T]) Fetch() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}

	index := s.Len() - 1
	element := s.data[index]
	return element, true
}

func (s *Stack[T]) ExFetch() T {
	value, ok := s.Fetch()
	if !ok {
		log.Fatal("Error: Fetch() from empty Stack")
	}
	return value
}

func (s *Stack[T]) Reverse() {
	half := len(s.data) / 2
	for i, j := 0, len(s.data)-1; i < half; i, j = i+1, j-1 {
		s.data[i], s.data[j] = s.data[j], s.data[i]
	}
}

func (s *Stack[T]) Reset() {
	s.data = s.data[:0]
}

func (s *Stack[T]) Values() iter.Seq[T] {
	return slices.Values(s.data)
}

func (s *Stack[T]) Backward() iter.Seq2[int, T] {
	return slices.Backward(s.data)
}

func (s *Stack[T]) All() iter.Seq2[int, T] {
	return slices.All(s.data)
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

type StackIter[T comparable] struct {
	stack *Stack[T]
	len   int
	index int
}

func (s *Stack[T]) Iter() *StackIter[T] {
	return &StackIter[T]{stack: s, len: s.Len(), index: -1}
}

func (s *StackIter[T]) Next() bool {
	if s.index < s.len-1 {
		s.index++
		return true
	}

	return false
}

func (s *StackIter[T]) Get() T {
	return s.stack.data[s.index]
}

func (s *Stack[T]) ContainsAny(values []T) bool {
	return slices.ContainsFunc(values, s.Contains)
}

func (s *Stack[T]) Contains(val T) bool {
	return slices.Contains(s.data, val)
}

func (s *Stack[T]) Index(val T) int {
	return slices.Index(s.data, val)
}

/*
func (s *Stack[T]) Each(f func(value T)) {
	for iter := s.Iter(); iter.Next(); {
		f(iter.Get())
	}
}
*/

// -------------------- SliceStack ----------------------------

type SliceStack[T comparable] []*Stack[T]

func (ss *SliceStack[T]) Contains(val T) bool {
	for _, i := range *ss {
		if i.Contains(val) {
			return true
		}
	}

	return false
}

func (ss *SliceStack[T]) Len() int {
	return len(*ss)
}

func (ss *SliceStack[T]) IsEmpty() bool {
	return ss.Len() == 0
}

func (ss *SliceStack[T]) Push(stk *Stack[T]) {
	*ss = append(*ss, stk)
}

func (ss *SliceStack[T]) Pop() (*Stack[T], bool) {
	if ss.IsEmpty() {
		return nil, false
	}

	index := ss.Len() - 1
	element := (*ss)[index]
	(*ss)[index] = nil
	*ss = (*ss)[:index]
	return element, true
}

func (ss *SliceStack[T]) ExPop() *Stack[T] {
	value, ok := ss.Pop()
	if !ok {
		log.Fatal("Error: Pop() from empty SliceStack")
	}
	return value
}
