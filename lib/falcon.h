// Copyright (c) 2024 The Falcon Contributors
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

#ifndef _FALCON_H_
#define _FALCON_H_

#include <stdint.h>
#include <stdbool.h>

typedef uintptr_t* yptr;
typedef int8_t ybyte;
typedef int16_t yshort;
typedef int8_t ychar;
typedef int32_t yint;
typedef int64_t ylong;
typedef float yfloat;
typedef double ydouble;
typedef bool ybool;

typedef struct {
    // Immutable data
    ychar* data;
    yint size;
} ystring;

typedef struct {
    yptr data;
    yint size;
} yarray;

#endif