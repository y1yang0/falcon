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
#include "falcon.h"
#include <stdio.h>
#include <stdlib.h>

#define DEBUG 0

// -----------------------------------------------------------------------------
// Runtime stubs for the compiler


yint* runtime_new_array(yint size) {
    yint* arr = (yint*)malloc(sizeof(yint)* size);
    if (DEBUG) {
        printf("++new_array: %d %p\n", size, arr);
    }
    return arr;
}

ystring* runtime_new_string(ychar* str, yint size) {
    ystring* s = (ystring*)malloc(sizeof(ystring));
    s->size = size;
    s->data = (ychar*)malloc(sizeof(ychar) * size);
    for (int i = 0; i < size; i++) {
        s->data[i] = str[i];
    }
    return s;
}

ystring* runtime_string_concat(ystring* a, ystring* b){
    ystring* s = (ystring*)malloc(sizeof(ystring));
    s->size = a->size + b->size;
    s->data = (ychar*)malloc(sizeof(ychar) * s->size);
    for (int i = 0; i < a->size; i++) {
        s->data[i] = a->data[i];
    }
    for (int i = 0; i < b->size; i++) {
        s->data[i + a->size] = b->data[i];
    }
    return s;
}

ybool runtime_string_eq(ystring* a, ystring* b){
    return a->size == b->size && memcmp(a->data, b->data, a->size) == 0;
}

ybool runtime_string_ne(ystring* a, ystring* b){
    return a->size != b->size || memcmp(a->data, b->data, a->size) != 0;
}

ybool runtime_string_lt(ystring* a, ystring* b){
    return memcmp(a->data, b->data, a->size) < 0;
}

ybool runtime_string_gt(ystring* a, ystring* b){
    return memcmp(a->data, b->data, a->size) > 0;
}

ybool runtime_string_le(ystring* a, ystring* b){
    return memcmp(a->data, b->data, a->size) <= 0;
}

ybool runtime_string_ge(ystring* a, ystring* b){
    return memcmp(a->data, b->data, a->size) >= 0;
}

yint runtime_string_cmp(ystring* a, ystring* b){
    return memcmp(a->data, b->data, a->size);
}

// -----------------------------------------------------------------------------
// Runtime Implementation


struct Heap {
    int* heap_base;
    int* heap_top;
};

struct Constants {
};

void runtime_init(){
    // TODO: Start GC Thread and periodically check for memory leaks
}


// The real program entry point
int entrypoint(int argc, char** argv) {
    if (DEBUG) {
        printf("++entrypoint\n");
    }
    runtime_init();
    __asm__("andq $-16, %rsp"); // Align stack to 16 bytes to follow the ABI
    __asm__("call main");
    exit(0);
}
