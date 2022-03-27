
: ? @ . ;
: -rot rot rot ;
: nip swap drop ;
: 2drop drop drop ;
: 2nip 2swap 2drop ;
: tuck swap over ;
: negate 0 swap - ;
: vswap
  { v1 v2 }
  v1 @ v2 @ v1 ! v2 !
;

: 0< 0 < ;
: 0= 0 = ;
: 0> 0 > ;
: 1+ 1 + ;
: 1- 1 - ;
: 2+ 2 + ;
: 2- 2 - ;
: 2* 2 * ;
: <> = not ;
: >= < not ;
: <= > not ;
: 0<> 0= not ;
: 0<= 0 <= ;
: 0>= 0 >= ;
: slow_mod begin 2dup >= while dup -rot - swap repeat drop ;
: mod 1 sys ;
: factor mod 0= ;
: even 2 factor ;

: min 2dup < if drop else nip then ;
: max 2dup > if drop else nip then ;

: false 0 ;
: true false not ;

: cell 1 ;
: cell+ 1+ ;
: th + ;
: ?+ dup ? cell+ ;
: +! dup @ rot + swap ! ;
: -! dup @ rot - swap ! ;

: 2! tuck ! cell+ ! ;
: 2@ dup cell+ @ swap @ ;
: ++ 1 swap +! ;
: -- 1 negate swap +! ;
: @+ dup cell+ swap @ ;

: newS 0 swap ! ;

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

: .s dup 1+ swap @ begin dup 0> while swap dup .c 1+ swap 1- repeat 2drop ;

: .s2 { pos }
  pos @
  pos 1+ to pos
  0 ?do
    pos dup .c 1+ to pos
  loop
;

: +s dup ++ dup @ + ! ;
: -s -- ;

: abs
  dup
  0< if
    negate
  then
;

: fak
  1 { x }
  1+ 1 ?do
    x i * to x
  loop
  x
;

: raise
  1 { s x n }
  begin
    n 0>
  while
    s x * to s
    n 1- to n
  repeat
  s
;

: toChar 65 + emit ;
: colon 58 emit ;
: showAlphabetPerChar
  26 0 do
    i toChar space colon space
    26 0 do
      i toChar
    loop
    cr
  loop
;

: integers 1+ 1 ?do i . space loop ;

