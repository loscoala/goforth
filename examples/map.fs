#!../goforth -file

: add2 2 + ;
: array 10 ;
: test
  array newS
  20 array +s
  40 array +s

  array &add2 map
  ??
;

: main test ;

