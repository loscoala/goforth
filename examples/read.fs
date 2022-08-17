#!../goforth -file

\ this is our base address in memory
: adr 0 ;

\ this is the size how much we want to read
: size 99 ;

( * This program reads from stdin and prints
  * the buffer to stdout.
  *
  * How to use it:
  * cat ../main.go | ./read.fs
)
: main
  begin
    adr size read
    adr @ if
      adr .s
    else
      leave
    then
  again
;

