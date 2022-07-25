# goforth

This is a compiler and interpreter for a forth-like language.

In general, Forth is, among other things, a programming language whose ideals are freedom and simplicity. Freedom requires knowledge and responsibility, and simplicity is yet more demanding. But "simple" can also mean rudimentary. This Forth attempts to be simple as possible and also fast as possible.

In this forth you can define local variables with curley brackets like in other forths and produce callable subroutines. The compiler produces human readable bytecode and is amazingly fast...

## Overview

`goforth` is the interactive compiler which produces a sequence of commands for a virtual stack machine also known as "Forth machine". The code is then interpreted and executed.

## Why goforth?

This implementation uses static memory at the runtime of the ForthVM. No allocations.

Memory needed:
1000x int64 <- for the memory
100x int64  <- locals
100x int64  <- return stack

Currently there is no parallel ForthVM execution via goroutines. If you want to have it, simply implement it into `sys()`.

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

## Installation

If you do want to get the latest version of `goforth`, change to any directory that is both outside of your `GOPATH` and outside of a module (a temp directory is fine), and run:

```sh
go install github.com/loscoala/goforth@latest
```

## Build and Dependencies

- All you need is a golang compiler 1.18. No other dependencies.

- Simply build the project with

```sh
git clone https://github.com/loscoala/goforth.git
cd goforth
go build
```

## Howto start

Start `./goforth` and in the REPL you can look what the dictionary contains by typing:

| command | description |
|---|---|
| % | Shows the complete dictionary |
| % name | Shows the definition of *name* in the dictionary |
| # filename | Loads a file |
| $ | Shows the values of the stack |

## Examples

Start the interpreter:

```sh
./goforth
```
Then inside the REPL type the following:

```forth
# examples/mandelbrot.fs   \ loads the file mandelbrot.fs and parses it
mb-init                    \ compile and run mb-init
mb                         \ compile and run mb
```

As a result you see a zoom in the mandelbrot fractal.

You can also define new words:

```forth
: myadd 2 + ;
```

Or words with a local context:

```forth
: fakulty
  1 { x }
  1+ 1 ?do
    x i * to x
  loop
  x
;
```

then run it:

```forth
5 fakulty .
```

which prints:

```forth
120
```

## Description of the files

| filename | description |
|---|---|
| main.go | The main function |
| compiler.go | The forth compiler |
| vm.go | A stack machine with additional functionality for forth |
| stack.go | A stack of strings implementation |
| label.go | A label generator for the bytecode |
| show.go | Visual representation and the REPL  |

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
| #id NOP | Does nothing. Used for labeling. | 
| JMP #id | Unconditional jump to label with id |
| JIN #id | Conditional jump to label with id if the top value on the stack is zero. |
| ADI | Simple addition of the top pair. |
| SBI | Simple subtraction of the top pair. |
| DVI | Simple division of the top pair. |
| LSI | 1 on the stack if the top element is less than the second element. 0 otherwise. |
| GRI | 1 on the stack if the top element id greater than the second element. 0 otherwise. |
| MLI | Multiplies the first two elements on the stack. |
| ADF | Simple float addition of the top pair. |
| SBF | Simple float subtraction of the top pair. |
| MLF | Multiplies the first two float elements on the stack. |
| DVF | Simple float division of the top pair. |
| PRF | Prints a float value from the stack as a number. |
| LSF | 1 on the stack if the top float element is less than the second element. 0 otherwise. |
| GRF | 1 on the stack if the top float element is greater than the second element. 0 otherwise. |
| OR | 1 if one of the first two elements on the stack unequal to 0 else 0. |
| AND | 1 if the first two elements on the stack unequal to 0 else 0. |
| NOT | 1 if the first element on the stack is 0 else 0.  |
| EQI | 1 if the first two element on the stack are equal else 0. |
| LV | Loads a value from the given address.  |
| L number | Loads a value on the stack. |
| LF float | Loads a float on the stack. |
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
| REF name | Pushes the address of a SUB on top of the stack |
| EXC | Pops the top value from the stack and calls a SUB routine. |
| PCK | dup = 0 pick |
| NRT | -rot = rot rot |
