
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
#include <math.h>

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

void rt_assert_char(ychar a, ychar b) {
    if(a != b){
        printf("Assertion failed: %c != %c\n", a, b);
        exit(1);
    }
}

void rt_assert_short(yshort a, yshort b){
    if(a != b){
        printf("Assertion failed: %d != %d\n", a, b);
        exit(1);
    }
}

void rt_assert_long(ylong a, ylong b){
    if(a != b){
        printf("Assertion failed: %ld != %ld\n", a, b);
        exit(1);
    }
}

void rt_assert_string(ystring* a, ystring* b){
    if(a->size != b->size){
        printf("Assertion failed: %d != %d\n", a->size, b->size);
        exit(1);
    }
    for(int i = 0; i < a->size; i++){
        if(a->data[i] != b->data[i]){
            printf("Assertion failed: %c != %c\n", a->data[i], b->data[i]);
            exit(1);
        }
    }
}

void rt_assert_double(ydouble a, ydouble b){
    ydouble epsilon = 0.000001; // close enough
    if(fabs(a - b) > epsilon){
        printf("Assertion failed: %lf != %lf\n", a, b);
        exit(1);
    }
}

ystring* rt_append(ystring* a, ychar c) {
    ystring* s = (ystring*)malloc(sizeof(ystring));
    s->size = a->size + 1;
    s->data = (ychar*)malloc(sizeof(ychar) * s->size);
    for (int i = 0; i < a->size; i++) {
        s->data[i] = a->data[i];
    }
    s->data[a->size] = c;
    return s;
}

void rt_cprint_char(ychar c){
    printf("%c\n", c);
}

void rt_cprint_arr(yint* arr, yint size){
    for(int i = 0; i < size; i++){
        printf("%d ", arr[i]);
    }
    printf("\n");
}

void rt_cprint_string(ystring* str){
    for(int i = 0; i < str->size; i++){
        printf("%c", str->data[i]);
    }
    printf("\n");
}

void rt_cprint_double(ydouble d){
    printf("%lf\n", d);
}