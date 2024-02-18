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

type BitMap struct {
	data []uint8
	size int
}

func NewBitMap(size int) *BitMap {
	return &BitMap{
		data: make([]uint8, (size+7)/8),
		size: size,
	}
}

func (bm *BitMap) Size() int {
	return bm.size
}

func (bm *BitMap) Set(i int) {
	ei := i / 8
	bm.data[ei] = bm.data[ei] | (1 << uint8(i%8))
}

func (bm *BitMap) Reset(i int) {
	ei := i / 8
	bm.data[ei] = bm.data[ei] & (^(1 << uint8(i%8)))
}

func (bm *BitMap) IsSet(i int) bool {
	return (bm.data[i/8] & (1 << uint8(i%8))) != uint8(0)
}

func (bm *BitMap) Unite(o *BitMap) bool {
	Assert(bm.size == o.size, "sanity check")
	l := len(bm.data)
	changed := false
	for i := 0; i < l; i++ {
		nv := bm.data[i] | o.data[i]
		if nv != bm.data[i] {
			bm.data[i] = nv
			changed = true
		}
	}
	return changed
}

func (bm *BitMap) Intersect(o *BitMap) bool {
	Assert(bm.size == o.size, "sanity check")
	l := len(bm.data)
	changed := false
	for i := 0; i < l; i++ {
		v := bm.data[i] & o.data[i]
		if v != bm.data[i] {
			bm.data[i] = v
			changed = true
		}
	}
	return changed
}

func (bm *BitMap) SetFrom(o *BitMap) bool {
	Assert(bm.size == o.size, "sanity check")
	changed := false
	for i := 0; i < len(o.data); i++ {
		if o.data[i] != bm.data[i] {
			bm.data[i] = o.data[i]
			changed = true
		}
	}
	return changed
}

func (bm *BitMap) Remove(o *BitMap) bool {
	Assert(bm.size == o.size, "sanity check")
	changed := false
	for i := 0; i < len(o.data); i++ {
		nv := bm.data[i] & (^o.data[i])
		if nv != bm.data[i] {
			bm.data[i] = nv
			changed = true
		}
	}
	return changed
}

func (bm *BitMap) Copy() *BitMap {
	newData := make([]uint8, len(bm.data))
	copy(newData, bm.data)
	return &BitMap{
		data: newData,
		size: bm.size,
	}
}
