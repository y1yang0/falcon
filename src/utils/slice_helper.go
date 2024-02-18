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

func InsertAt[T any](slice []T, index int, e T) []T {
	// TODO: check index

	if index == len(slice) {
		return append(slice, e)
	}

	res := make([]T, len(slice)+1)
	copy(res[:index], slice[:index])
	res[index] = e
	copy(res[index+1:], slice[index:])

	return res
}
