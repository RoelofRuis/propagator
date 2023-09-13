package propagator

import (
	"fmt"
	"strings"
)

type Void struct{}

var Contained Void

type Set[A comparable] map[A]Void

func NewSet[A comparable](values ...A) Set[A] {
	set := make(map[A]Void)
	for _, v := range values {
		set[v] = Contained
	}
	return set
}

func (s Set[A]) Intersect(that Set[A]) Set[A] {
	result := NewSet[A]()
	for elem := range s {
		if that.Contains(elem) {
			result.Insert(elem)
		}
	}
	return result
}

func (s Set[A]) Insert(value A) Set[A] {
	set := s
	if set == nil {
		set = Set[A]{}
	}
	set[value] = Contained
	return set
}

func (s Set[A]) Contains(value A) bool {
	_, has := s[value]
	return has
}

func (s Set[A]) ContainsOneOf(values []A) bool {
	for _, candidate := range values {
		if s.Contains(candidate) {
			return true
		}
	}
	return false
}

func (s Set[A]) Size() int {
	return len(s)
}

func (s Set[A]) IsEmpty() bool {
	return s.Size() == 0
}

func (s Set[A]) String() string {
	var res []string
	for e := range s {
		res = append(res, fmt.Sprintf("%v", e))
	}
	return fmt.Sprintf("Set[%s]", strings.Join(res, " "))
}
