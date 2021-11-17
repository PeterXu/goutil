package util

import "sync"

type Empty struct{}

var empty Empty

type Set struct {
	sync.Mutex
	m map[interface{}]Empty
}

func NewSet(items ...interface{}) *Set {
	s := &Set{
		m: make(map[interface{}]Empty),
	}
	s.Add(items...)
	return s
}

func (s *Set) Add(items ...interface{}) {
	s.Lock()
	defer s.Unlock()

	for _, item := range items {
		s.m[item] = empty
	}
}

func (s *Set) Contain(item interface{}) bool {
	s.Lock()
	defer s.Unlock()

	_, ok := s.m[item]
	return ok
}

func (s *Set) Remove(item interface{}) {
	s.Lock()
	defer s.Unlock()

	delete(s.m, item)
}

func (s *Set) Empty() bool {
	s.Lock()
	defer s.Unlock()

	return (len(s.m) == 0)
}

func (s *Set) Len() int {
	s.Lock()
	defer s.Unlock()

	return len(s.m)
}

func (s *Set) Clear() {
	s.Lock()
	defer s.Unlock()

	s.m = make(map[interface{}]Empty)
}

func (s *Set) Values() []interface{} {
	s.Lock()
	defer s.Unlock()

	var values []interface{}
	for key, _ := range s.m {
		values = append(values, key)
	}
	return values
}
