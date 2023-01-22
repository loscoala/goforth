# goforth

This is a compiler and byte code interpreter for a forth-like language.

In general, Forth is, among other things, a programming language whose ideals are freedom and simplicity. Freedom requires knowledge and responsibility, and simplicity is yet more demanding. But "simple" can also mean rudimentary. This Forth attempts to be simple as possible and also fast as possible.

In this forth you can define local variables with curley brackets like in other forths and produce callable byte code subroutines. The compiler produces human readable bytecode and is amazingly fast.

## Overview

`goforth` is the interactive compiler which produces a sequence of commands for a virtual stack machine also known as "Forth machine". The code is then interpreted and executed.

## Why goforth?

Goforth can be used as an embeddable programming language. See topic Embedding.

Currently there is no parallel ForthVM execution via goroutines. If you want to have it, simply implement it into `sys()`. Maybe in the future.

## Usage

Simply execute `goforth`. See "Installation"

1. You can also execute forth-scripts:

```sh
goforth --file=forthscript.fs
```

```sh
echo ": main 5 5 + . ;" | goforth
```

2. In the Shebang you can alternatively write the following:

```md
#!goforth --file
# code goes here...
```

3. Or in the command line:

```sh
goforth -script '
: myfunc ." Hello World" ;
: main myfunc ;
'
```

The compiler has `core.fs` automatically included into the binary.

## Installation

If you do want to get the latest version of `goforth`, change to any directory that is both outside of your `GOPATH` and outside of a module (a temp directory is fine), and run:

```sh
go install github.com/loscoala/goforth/cmd/goforth@latest
```

If you do want to add the latest version to your go.mod inside your project run:

```sh
go get github.com/loscoala/goforth@latest
```

## Build and Dependencies

- All you need is a golang compiler 1.19.

- Simply build the project with build.sh

- For C code generation you need a working gcc-12 compiler.

```sh
git clone https://github.com/loscoala/goforth.git
./build.sh
```

## Howto start

Start `./goforth` and in the REPL you can look what the dictionary contains by typing:

| command | description |
|---|---|
| % | Shows the complete dictionary |
| % name | Shows the definition of *name* in the dictionary |
| use filename | Loads a file |
| $ | Shows the values of the stack |

## Examples

### Mandelbrot

Start the interpreter:

```sh
./cmd/goforth/goforth
```
Then inside the REPL type the following:

```forth
forth> use examples/mandelbrot.fs \ loads the file mandelbrot.fs and parses it
forth> mb-init                    \ compile and run mb-init
forth> true debug                 \ OPTIONAL in order to show byte code and benchmark
forth> mb                         \ compile and run mb
```

As a result you see a zoom in the mandelbrot fractal.

![mandelbrot-video](https://github.com/loscoala/goforth/raw/main/examples/manbelbrot.gif)

### Interactive development

You can also define new words:

```forth
forth> : myadd 2 + ;
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
forth> 5 fakulty .
```

which prints:

```forth
120
```

### Generate C-Code

To translate one or more words into C code and generate a native binary file there is a `compile` statement.

This is how the `html.fs` example written in Forth can be easily translated into C and executed immediately:

```forth
forth> use examples/html.fs
forth> compile test
```

The word `test` was defined in the sample file as follows:

```forth
: test
  document
  [
    [
      [ ." charset=\"utf-8\"" ] meta
      [ ." Example Page" ] title
    ] head
    [
      [ ." Example Page" ] h1
      [ ." Hello " [ ." World!" ] b ] p
    ] body
  ] html
;
```
After the `compile test` statement, a C file `main.c` was generated in the `lib` directory and compiled with the gcc-11 compiler and executed.

The result is also shown as follows:

```html
<!doctype html>
<html lang="de-DE"><head><meta charset="utf-8"><title>Example Page</title></head><body><h1>Example Page</h1><p>Hello <b>World!</b></p></body></html>
```

### Debugging

By calling `true debug` you can enable the benchmark mode.

```forth
forth> true debug
```

Now the byte code is displayed and the execution time of the program.

```
forth> true debug
forth> 5 3 min .
SUB min;TDP;LSI;JIN #0;DRP;JMP #1;#0 NOP;SWP;DRP;#1 NOP;END;
MAIN;L 5;L 3;CALL min;PRI;STP;
3

execution time: 15.947µs
Number of Cmds: 13
Speed: 0.000815 cmd/ns
```

The actual debugger can be run like this:

```forth
forth> debug 34 21 min .
```

Which gives the following result:

```
SUB min;TDP;LSI;JIN #0;DRP;JMP #1;#0 NOP;SWP;DRP;#1 NOP;END;
MAIN;L 34;L 21;CALL min;PRI;STP;

 11 MAIN            |                           |                           |
 12 L 34            | 34                        |                           |
 13 L 21            | 34 21                     |                           |
 14 CALL min        | 34 21                     | 14                        |
  1 TDP             | 34 21 34 21               | 14                        |
  2 LSI             | 34 21 0                   | 14                        |
  3 JIN #0          | 34 21                     | 14                        |
  6 NOP #0          | 34 21                     | 14                        |
  7 SWP             | 21 34                     | 14                        |
  8 DRP             | 21                        | 14                        |
  9 NOP #1          | 21                        | 14                        |
 10 END             | 21                        |                           |
 15 PRI             |                           |                           | 21
 16 STP
```

As you can see on the top there is the ByteCode and below you see the program pointer, the command, the stack, the return stack and the output.

## Embedding

First, you have to add goforth to go.mod:

```sh
go get github.com/loscoala/goforth@latest
```

Then in golang you can import goforth:

```go
import "github.com/loscoala/goforth"
```

Now all you need is a ForthCompiler:

```go
fc := goforth.NewForthCompiler()

fc.Fvm.Sysfunc = func(fvm *goforth.ForthVM, syscall int64) {
  switch syscall {
  case 999:
    value := fvm.Pop()
    fvm.Push(value + value)
    fmt.Println("This is a custom sys call in Forth")
  default:
    fmt.Println("Not implemented")
  }
}

// Parse the Core lib:
if err := fc.Parse(goforth.Core); err != nil {
  goforth.PrintError(err)
}

// Run some code:
if err := fc.Run(": main .\" Hello World!\" ;"); err != nil {
  goforth.PrintError(err)
}

// Call custom syscall (calculated 10+10 and prints it):
if err := fc.Run(": customcall 999 sys ; : main 10 customcall . ;"); err != nil {
  goforth.PrintError(err)
}
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
| config.go | Configuration during runtime |

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
| TR | to r |
| FR | from r |
| RF | r fetch |
| TTR | 2 to r |
| TFR | 2 from r |
| TRF | 2 r fetch |
