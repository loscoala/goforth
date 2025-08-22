: ? @ . ;
: ?? memsize 0 mem ;
: mem ?do i @ ?dup if i . colon . space then loop ;

: ls [ a" ls -lhA --color=always" shell ] alloc ;
: vim [ a" vim" shell ] alloc ;

\ iend is the upper limit inside a do .. loop
: iend 2r> 2dup 2>r drop ;
: i r@ ;
: j    \ jend j iend i jx --
  r>   \ jend j iend i -- jx
  r>   \ jend j iend -- jx i
  r>   \ jend j -- jx i iend
  r@   \ jend j -- jx i iend j
  -rot \ jend j -- jx j i iend
  >r   \ jend j iend -- jx j i
  >r   \ jend j iend i -- jx j
  swap \ jend j iend i -- j jx
  >r   \ jend j iend i jx -- j
;

: 2! tuck ! cell+ ! ;
: 2@ dup cell+ @ swap @ ;
: ++ dup @ 1+ swap ! ;
: -- dup @ 1- swap ! ;
: @+ dup cell+ swap @ ;

: squared dup * ;
: cubed dup squared * ;
: 4th squared squared ;

: cr 10 emit ;
: quo 34 emit ;
: bl 32 ;
: space bl emit ;
: spaces
  begin
  dup
  0> while
    space 1-
  repeat
  drop
;

: lb 40 emit ;
: rb 41 emit ;
: .c @ emit ;

: .s sv:print ;
: print drop begin dup 0> while emit repeat drop ;

: toChar 65 + emit ;
: colon 58 emit ;
