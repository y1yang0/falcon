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
let y = 20

// or omit type and let the compiler infer it
z = 30
```

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

