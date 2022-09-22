#!goforth -file

: add2 2 + ;
: array 10 ;
: init
  20 allocate
  array newS
  20 array +s
  40 array +s
;
: test
  init
  array &add2 map
  ??
;
: test2
  init
  array [ 2 + ] map
  ??
;

: main test cr test2 ;

