# goforth

This is a compiler and interpreter for a forth-like language.

In general, Forth is, among other things, a programming language whose ideals are freedom and simplicity. Freedom requires knowledge and responsibility, and simplicity is yet more demanding. But "simple" can also mean rudimentary. This Forth attempts to be simple as possible and also fast as possible.

In this forth you can define local variables with curley brackets like in other forths and produce callable subroutines. The compiler produces human readable bytecode and is amazingly fast..

## Overview

`goforth` is the interactive compiler which produces a sequence of commands for a virtual stack machine also known as "Forth machine". The code is then interpreted and executed.

## Usage

Simply execute `goforth`

You can also execute forth-scripts:

```sh
goforth --file=forthscript.fs
```

In the Shebang you can alternatively write the following:

```md
#!goforth --file
# code goes here...
```

The compiler loads `core.fs` automatically.

## Dependencies

- All you need is a golang compiler 1.17.x. No other dependencies.

## Description of the files

| filename | description |
|---|---|
| main.go | The REPL and main function |
| compiler.go | The forth compiler |
| vm.go | A stack machine with additional functionality for forth |
| stack.go | A stack of strings implementation |
| label.go | A label generator for the bytecode |
| show.go | TODO |

## VM commands

The stack machine consists of a single LIFO stack and memory. Memory is adressed by an unsigned number. The stack and memory consists only of numbers.

| command | description |
|---|---|
| RDI | Reads a value from the input. The input is implementation dependant. |
| PRI | Prints a value from the stack as a number. |
| PRA | Prints a value from the stack as an character. |
| DUP | Duplicates the top stack value. |
| OVR | Copies the second value from the top on the top. |
| TVR | Copies the second pair onto the top pair. |
| TWP | Swaps the top pair with the pair under it. |
| QDP | ?dup implemented as: if the top value is zero - duplicate it - otherwise do nothing. |
| ROT | Rotates the top three elements by pushing to stack one element "down" an taking the last element on the top. |
| TDP | Duplicates the top pair on the stack. |
| DRP | Drops the top element on the stack. |
| SWP | Swaps the top pair on the stack.  |
| ADI | Simple addition of the top pair. |
| #id NOP | Does nothing. Used for labeling. | 
| JMP #id | Unconditional jump to label with id |
| JIN #id | Conditional jump to label with id if the top value on the stack is zero. |
| SBI | Simple sustraction of the top pair. |
| DVI | Simple division of the top pair. |
| LSI | 1 on the stack if the top element is less than the second element. 0 otherwise.|
| GRI | 1 on the stack if the top element id greater than the second element. 0 otherwise. |
| MLI | Multiplies the first two elements on the stack. |
| OR | 1 if one of the first two elements on the stack unequal to 0 else 0. |
| AND | 1 if the first two elements on the stack unequal to 0 else 0. |
| NOT | 1 if the first element on the stack is 0 else 0.  |
| EQI | 1 if the first two element on the stack are equal else 0. |
| LV value | Push the value on top of the stack.  |
| L number | Loads a value of the stack. |
| STR | Stores the second element from the stack into the memory address which is the first element from the stack. |
| SYS | Make a syscall. The top element from the stack is used to make different syscalls. |
| STP | Quits the execution. |
| SUB name | Declares a subroutine. Used only for code generation. Not used by the vm |
| END | Ends the subroutine. Used only for code generation. Not used by the vm |
| MAIN | Declares the beginning of the main. Used only for code generation. Not used by the vm |
| LCTX | Creates a new context of local variables  |
| LDEF | Pops the top value of the stack and copies it to the local definitions |
| LSET | Assigns the top value of the stack to a local variable  |
| LCL | Pushes the local value on top of the stack |
| LCLR | Clears the local definitions |
| CALL name | Call a SUB routine. |
