package util

import (
	"sort"
)

type empty struct{}

// StringSet is a set of strings, implemented via map[string]struct{} for minimal memory consumption.
type StringSet map[string]empty

// NewStringSet creates a StringSet from a list of values.
func NewStringSet(items ...string) StringSet {
	ss := StringSet{}
	ss.Insert(items...)
	return ss
}

// Insert adds items to the set.
func (s StringSet) Insert(items ...string) {
	for _, item := range items {
		s[item] = empty{}
	}
}

// Delete removes item from the set.
func (s StringSet) Delete(item string) {
	delete(s, item)
}

// Has returns true if item is contained in the set.
func (s StringSet) Has(item string) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true if all items are contained in the set.
func (s StringSet) HasAll(items ...string) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// IsSuperset returns true if s1 is a supreset of s2.
func (s StringSet) IsSuperset(s2 StringSet) bool {
	for item := range s2 {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// List returns the contents as a sorted string slice.
func (s StringSet) List() []string {
	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Strings(res)
	return res
}
