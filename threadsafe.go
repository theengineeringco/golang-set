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

import "sync"

type Set[T comparable] struct {
	sync.RWMutex
	uss ThreadUnsafeSet[T]
}

// NewSet creates and returns a new set with the given elements.
// Operations on the resulting set are thread-safe.
func NewSet[T comparable](vals ...T) *Set[T] {
	s := newSet[T]()
	for _, item := range vals {
		s.Add(item)
	}
	return &s
}

func newSet[T comparable]() Set[T] {
	newUss := newThreadUnsafeSet[T]()
	return Set[T]{
		uss: newUss,
	}
}

func (s *Set[T]) Add(v T) bool {
	s.Lock()
	ret := s.uss.Add(v)
	s.Unlock()
	return ret
}

func (s *Set[T]) Contains(v ...T) bool {
	s.RLock()
	ret := s.uss.Contains(v...)
	s.RUnlock()
	return ret
}

func (s *Set[T]) IsSubset(other SetInterface[T]) bool {
	o := other.(*Set[T])

	s.RLock()
	o.RLock()

	ret := s.uss.IsSubset(&o.uss)
	s.RUnlock()
	o.RUnlock()
	return ret
}

func (s *Set[T]) IsProperSubset(other SetInterface[T]) bool {
	o := other.(*Set[T])

	s.RLock()
	defer s.RUnlock()
	o.RLock()
	defer o.RUnlock()

	return s.uss.IsProperSubset(&o.uss)
}

func (s *Set[T]) IsSuperset(other SetInterface[T]) bool {
	return other.IsSubset(s)
}

func (s *Set[T]) IsProperSuperset(other SetInterface[T]) bool {
	return other.IsProperSubset(s)
}

func (s *Set[T]) Union(other SetInterface[T]) SetInterface[T] {
	o := other.(*Set[T])

	s.RLock()
	o.RLock()

	unsafeUnion := s.uss.Union(&o.uss).(*ThreadUnsafeSet[T])
	ret := &Set[T]{uss: *unsafeUnion}
	s.RUnlock()
	o.RUnlock()
	return ret
}

func (s *Set[T]) Intersect(other SetInterface[T]) SetInterface[T] {
	o := other.(*Set[T])

	s.RLock()
	o.RLock()

	unsafeIntersection := s.uss.Intersect(&o.uss).(*ThreadUnsafeSet[T])
	ret := &Set[T]{uss: *unsafeIntersection}
	s.RUnlock()
	o.RUnlock()
	return ret
}

func (s *Set[T]) Difference(other SetInterface[T]) SetInterface[T] {
	o := other.(*Set[T])

	s.RLock()
	o.RLock()

	unsafeDifference := s.uss.Difference(&o.uss).(*ThreadUnsafeSet[T])
	ret := &Set[T]{uss: *unsafeDifference}
	s.RUnlock()
	o.RUnlock()
	return ret
}

func (s *Set[T]) SymmetricDifference(other SetInterface[T]) SetInterface[T] {
	o := other.(*Set[T])

	s.RLock()
	o.RLock()

	unsafeDifference := s.uss.SymmetricDifference(&o.uss).(*ThreadUnsafeSet[T])
	ret := &Set[T]{uss: *unsafeDifference}
	s.RUnlock()
	o.RUnlock()
	return ret
}

func (s *Set[T]) Clear() {
	s.Lock()
	s.uss = newThreadUnsafeSet[T]()
	s.Unlock()
}

func (s *Set[T]) Remove(v T) {
	s.Lock()
	delete(s.uss, v)
	s.Unlock()
}

func (s *Set[T]) Cardinality() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.uss)
}

func (s *Set[T]) Each(cb func(T) bool) {
	s.RLock()
	for elem := range s.uss {
		if cb(elem) {
			break
		}
	}
	s.RUnlock()
}

func (s *Set[T]) Iter() <-chan T {
	ch := make(chan T)
	go func() {
		s.RLock()

		for elem := range s.uss {
			ch <- elem
		}
		close(ch)
		s.RUnlock()
	}()

	return ch
}

func (s *Set[T]) Iterator() *Iterator[T] {
	iterator, ch, stopCh := newIterator[T]()

	go func() {
		s.RLock()
	L:
		for elem := range s.uss {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
		s.RUnlock()
	}()

	return iterator
}

func (s *Set[T]) Equal(other SetInterface[T]) bool {
	o := other.(*Set[T])

	s.RLock()
	o.RLock()

	ret := s.uss.Equal(&o.uss)
	s.RUnlock()
	o.RUnlock()
	return ret
}

func (s *Set[T]) Clone() SetInterface[T] {
	s.RLock()

	unsafeClone := s.uss.Clone().(*ThreadUnsafeSet[T])
	ret := &Set[T]{uss: *unsafeClone}
	s.RUnlock()
	return ret
}

func (s *Set[T]) String() string {
	s.RLock()
	ret := s.uss.String()
	s.RUnlock()
	return ret
}

func (s *Set[T]) Pop() (T, bool) {
	s.Lock()
	defer s.Unlock()
	return s.uss.Pop()
}

func (s *Set[T]) ToSlice() []T {
	keys := make([]T, 0, s.Cardinality())
	s.RLock()
	for elem := range s.uss {
		keys = append(keys, elem)
	}
	s.RUnlock()
	return keys
}

func (s *Set[T]) MarshalJSON() ([]byte, error) {
	s.RLock()
	b, err := s.uss.MarshalJSON()
	s.RUnlock()

	return b, err
}

func (s *Set[T]) UnmarshalJSON(p []byte) error {
	s.RLock()
	err := s.uss.UnmarshalJSON(p)
	s.RUnlock()

	return err
}
