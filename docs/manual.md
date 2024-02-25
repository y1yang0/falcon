# The Falcon Programming Language Reference

# Types
Falcon supports the following types:
- int: 32-bit signed integer, range from -2,147,483,648 to 2,147,483,647
- long: 64-bit signed integer, range from -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807
- short: 16-bit signed integer, range from -32,768 to 32,767
- float: 32-bit floating point number
- double: 64-bit floating point number
- char: 8-bit ascii character
- bool: boolean value
- byte: 8-bit signed integer, range from -128 to 127
- string: sequence of characters
- void: no value
- array: sequence of elements of the same type

# Variables
Variables are used to store data. They are declared using the let keyword followed by the letiable name and the type of the letiable. The type of the letiable is optional, if not specified, the type is inferred from the value assigned to the letiable.

```falcon
let x int = 10
let x int = 10
let y long = 10L
let z float = 10.0F
let a double = 10.0
let b char = 'a'
let c bool = true
let d byte = 10B
let e short = 10S

// or omit type and let the compiler infer it
z = 30
```
The letter followed by the type is a suffix to specify the type of the variable explicitly. For example, 10L is a long, 10.0F is a float, 10.0 is a double, 'a' is a char, true is a bool, and 10B is a byte.

# Functions
Functions are used to group a set of statements together to perform a specific task. Functions are declared using the func keyword followed by the function name, a list of parameters and the return type of the function.

```falcon
func add(x int, y int) int {
    return x + y
}
```

# Control Flow
Falcon supports the following control flow statements
## 1. If-then-else
```falcon
if x > 10 {
    print("x is greater than 10")
} else {
    print("x is less than or equal to 10")
}
```
Or even more fancy:
```falcon
let x int = 10
if x > 10 {
    print("x is greater than 10")
} while x < 10 {
    print("x is less than 10")
    x -= 1
}
```

## 2. For loop
```falcon
for i = 0; i < 10; i+=1 {
    print(i)
}
```

## 3. While loop
```falcon
i = 0
while i < 10 {
    print(i)
    i+=1
}
```

## Break
```falcon
while true {
    break
}
```

## Continue
```falcon
for i = 0; i < 10; i+=1 {
    if i == 5 {
        continue
    }
    print(i)
}
```

# Operators
Falcon supports the following operators
- Arithmetic operators: `+, -, *, /, %`
- Relational operators: `==, !=, <, >, <=, >=`
- Logical operators: `&&, ||, !`
- Bitwise operators: `&, |, ^, <<, >>`
- Assignment operators: `=, +=, -=, *=, /=, %=`
- Ternary operator: `?:`

# Comments
Falcon supports single line and multi-line comments
```falcon
// This is a single line comment
```

