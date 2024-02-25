
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

// Test purpose, temporary, will remove later
void rt_cprint_arr(yint* arr, yint size){
    for(int i = 0; i < size; i++){
        printf("%d ", arr[i]);
    }
    printf("\n");
}
