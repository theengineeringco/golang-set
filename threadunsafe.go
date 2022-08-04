/*
Open Source Initiative OSI - The MIT License (MIT):Licensing

The MIT License (MIT)
Copyright (c) 2013 - 2022 Ralph Caraveo (deckarep@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mapset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type ThreadUnsafeSet[T comparable] map[T]struct{}

// Assert concrete type:threadUnsafeSet adheres to Set interface.
var _ SetInterface[string] = (*ThreadUnsafeSet[string])(nil)

// NewThreadUnsafeSet creates and returns a new set with the given elements.
// Operations on the resulting set are not thread-safe.
func NewThreadUnsafeSet[T comparable](vals ...T) *ThreadUnsafeSet[T] {
	s := newThreadUnsafeSet[T]()
	for _, item := range vals {
		s.Add(item)
	}
	return &s
}

func newThreadUnsafeSet[T comparable]() ThreadUnsafeSet[T] {
	return make(ThreadUnsafeSet[T])
}

func (s *ThreadUnsafeSet[T]) Add(v T) bool {
	prevLen := len(*s)
	(*s)[v] = struct{}{}
	return prevLen != len(*s)
}

// private version of Add which doesn't return a value
func (s *ThreadUnsafeSet[T]) add(v T) {
	(*s)[v] = struct{}{}
}

func (s *ThreadUnsafeSet[T]) Cardinality() int {
	return len(*s)
}

func (s *ThreadUnsafeSet[T]) Clear() {
	*s = newThreadUnsafeSet[T]()
}

func (s *ThreadUnsafeSet[T]) Clone() SetInterface[T] {
	clonedSet := make(ThreadUnsafeSet[T], s.Cardinality())
	for elem := range *s {
		clonedSet.add(elem)
	}
	return &clonedSet
}

func (s *ThreadUnsafeSet[T]) Contains(v ...T) bool {
	for _, val := range v {
		if _, ok := (*s)[val]; !ok {
			return false
		}
	}
	return true
}

// private version of Contains for a single element v
func (s *ThreadUnsafeSet[T]) contains(v T) bool {
	_, ok := (*s)[v]
	return ok
}

func (s *ThreadUnsafeSet[T]) Difference(other SetInterface[T]) SetInterface[T] {
	o := other.(*ThreadUnsafeSet[T])

	diff := newThreadUnsafeSet[T]()
	for elem := range *s {
		if !o.contains(elem) {
			diff.add(elem)
		}
	}
	return &diff
}

func (s *ThreadUnsafeSet[T]) Each(cb func(T) bool) {
	for elem := range *s {
		if cb(elem) {
			break
		}
	}
}

func (s *ThreadUnsafeSet[T]) Equal(other SetInterface[T]) bool {
	o := other.(*ThreadUnsafeSet[T])

	if s.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range *s {
		if !o.contains(elem) {
			return false
		}
	}
	return true
}

func (s *ThreadUnsafeSet[T]) Intersect(other SetInterface[T]) SetInterface[T] {
	o := other.(*ThreadUnsafeSet[T])

	intersection := newThreadUnsafeSet[T]()
	// loop over smaller set
	if s.Cardinality() < other.Cardinality() {
		for elem := range *s {
			if o.contains(elem) {
				intersection.add(elem)
			}
		}
	} else {
		for elem := range *o {
			if s.contains(elem) {
				intersection.add(elem)
			}
		}
	}
	return &intersection
}

func (s *ThreadUnsafeSet[T]) IsProperSubset(other SetInterface[T]) bool {
	return s.Cardinality() < other.Cardinality() && s.IsSubset(other)
}

func (s *ThreadUnsafeSet[T]) IsProperSuperset(other SetInterface[T]) bool {
	return s.Cardinality() > other.Cardinality() && s.IsSuperset(other)
}

func (s *ThreadUnsafeSet[T]) IsSubset(other SetInterface[T]) bool {
	o := other.(*ThreadUnsafeSet[T])
	if s.Cardinality() > other.Cardinality() {
		return false
	}
	for elem := range *s {
		if !o.contains(elem) {
			return false
		}
	}
	return true
}

func (s *ThreadUnsafeSet[T]) IsSuperset(other SetInterface[T]) bool {
	return other.IsSubset(s)
}

func (s *ThreadUnsafeSet[T]) Iter() <-chan T {
	ch := make(chan T)
	go func() {
		for elem := range *s {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

func (s *ThreadUnsafeSet[T]) Iterator() *Iterator[T] {
	iterator, ch, stopCh := newIterator[T]()

	go func() {
	L:
		for elem := range *s {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
	}()

	return iterator
}

// TODO: how can we make this properly , return T but can't return nil.
func (s *ThreadUnsafeSet[T]) Pop() (v T, ok bool) {
	for item := range *s {
		delete(*s, item)
		return item, true
	}
	return
}

func (s *ThreadUnsafeSet[T]) Remove(v T) {
	delete(*s, v)
}

func (s *ThreadUnsafeSet[T]) String() string {
	items := make([]string, 0, len(*s))

	for elem := range *s {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(items, ", "))
}

func (s *ThreadUnsafeSet[T]) SymmetricDifference(other SetInterface[T]) SetInterface[T] {
	o := other.(*ThreadUnsafeSet[T])

	sd := newThreadUnsafeSet[T]()
	for elem := range *s {
		if !o.contains(elem) {
			sd.add(elem)
		}
	}
	for elem := range *o {
		if !s.contains(elem) {
			sd.add(elem)
		}
	}
	return &sd
}

func (s *ThreadUnsafeSet[T]) ToSlice() []T {
	keys := make([]T, 0, s.Cardinality())
	for elem := range *s {
		keys = append(keys, elem)
	}

	return keys
}

func (s *ThreadUnsafeSet[T]) Union(other SetInterface[T]) SetInterface[T] {
	o := other.(*ThreadUnsafeSet[T])

	n := s.Cardinality()
	if o.Cardinality() > n {
		n = o.Cardinality()
	}
	unionedSet := make(ThreadUnsafeSet[T], n)

	for elem := range *s {
		unionedSet.add(elem)
	}
	for elem := range *o {
		unionedSet.add(elem)
	}
	return &unionedSet
}

// MarshalJSON creates a JSON array from the set, it marshals all elements
func (s *ThreadUnsafeSet[T]) MarshalJSON() ([]byte, error) {
	items := make([]string, 0, s.Cardinality())

	for elem := range *s {
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}

		items = append(items, string(b))
	}

	return []byte(fmt.Sprintf("[%s]", strings.Join(items, ","))), nil
}

// UnmarshalJSON recreates a set from a JSON array, it only decodes
// primitive types. Numbers are decoded as json.Number.
func (s *ThreadUnsafeSet[T]) UnmarshalJSON(b []byte) error {
	var i []any

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	if *s == nil {
		*s = make(map[T]struct{}, len(i))
	}

	for _, v := range i {
		switch t := v.(type) {
		case T:
			s.add(t)
		default:
			// anything else must be skipped.
			continue
		}
	}

	return nil
}
