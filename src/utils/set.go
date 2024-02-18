// Copyright (c) 2024 The Sprite Programming Language
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

type Set[T any] struct {
	m map[any]any
}

func NewSet[T any]() *Set[T] {
	return &Set[T]{
		m: make(map[any]any),
	}
}

func (set *Set[T]) Add(e T) bool {
	if set.Contains(e) {
		return false
	}
	set.m[e] = nil
	return true
}

func (set *Set[T]) Remove(e T) bool {
	if !set.Contains(e) {
		return false
	}
	delete(set.m, e)
	return true
}

func (set *Set[T]) Contains(e T) bool {
	_, ok := set.m[e]
	return ok
}

func (set *Set[T]) Length() int {
	return len(set.m)
}

func (set *Set[T]) ForEach(f func(T)) {
	for k := range set.m {
		f(k.(T))
	}
}
