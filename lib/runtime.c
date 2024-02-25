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
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

#define DEBUG 0

typedef uintptr_t yptr;
typedef int8_t ybyte;
typedef int16_t yshort;
typedef int32_t yint;
typedef int64_t ylong;
typedef float yfloat;
typedef double ydouble;
typedef bool ybool;

// -----------------------------------------------------------------------------
// Standard Library Native Functions


void rt_cprint(yint n){
    printf("%d\n", n);
}

void rt_cprint_long(ylong n){
    printf("%ld\n", n);
}

void rt_cprint_bool(ybool n){
    printf("%s\n", n==1?"true":"false");
}

void rt_assert(yint a, yint b){
    if(a != b){
        printf("Assertion failed: %d != %d\n", a, b);
        exit(1);
    }
}

void rt_assert_bool(ybool a, ybool b){
    if(a != b){
        printf("Assertion failed: %d != %d\n", a, b);
        exit(1);
    }
}

// Test purpose, temporary, will remove later
void rt_cprint_arr(yint* arr, yint size){
    for(int i = 0; i < size; i++){
        printf("%d ", arr[i]);
    }
    printf("\n");
}

// -----------------------------------------------------------------------------
// Runtime stubs for the compiler


yint* runtime_new_array(yint size) {
    yint* arr = (yint*)malloc(sizeof(yint)* size);
    if (DEBUG) {
        printf("++new_array: %d %p\n", size, arr);
    }
    return arr;
}

yint* runtime_new_string(yint* str, yint size) {
    char* s = (char*)malloc(sizeof(yint)* size);
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
    __asm__("call main");
    exit(0);
}
