\ : -rot rot rot ;
: nip swap drop ;
: 2drop drop drop ;
: 2nip 2swap 2drop ;
: tuck swap over ;
: spin -rot swap ;
: negate 0 swap - ;
: fnegate 0. swap f- ;
: vswap
  { v1 v2 }
  v1 @ v2 @ v1 ! v2 !
;

: false 0 ;
: true 1 ;

: 0< 0 < ;
: 0= 0 = ;
: 0> 0 > ;
: 1+ inc ;
: 1- dec ;
: 2+ 2 + ;
: 2- 2 - ;
: 2* 2 * ;
: <> = not ;
: >= < not ;
: f>= f< not ;
: <= > not ;
: f<= f> not ;
: 0<> 0= not ;
: 0<= 0 <= ;
: 0>= 0 >= ;
: slow_mod begin 2dup >= while dup -rot - swap repeat drop ;
: depth 0 sys ;
: $ depth begin dup 0> while dup pick . space 1- repeat drop ;
: mod 1 sys ;
: fsqrt 2 sys ;
: factor mod 0= ;
: even 2 factor ;
: within ( u ul uh -- t ) >r over r> <= >r >= r> and ;

: min 2dup < if drop else nip then ;
: max 2dup > if drop else nip then ;

: cell 1 ;
: cell+ 1+ ;
: float 1 ;
: float+ 1+ ;
: th + ;
: ?+ dup ? cell+ ;
: +! dup @ rot + swap ! ;
: -! dup @ rot - swap ! ;

: z! swap dup ! float+ ! ;
: z@ dup @ float+ @ ;
: z. swap ." (" f. ." , " f. ." i)" ;
: zdup 2dup ;
: zdrop 2drop ;
: zover 2over ;
: zswap 2swap ;
: re drop ;
: im nip ;
: z= rot = -rot = and ;
: z+ rot f+ -rot f+ swap ;
: zabs dup f* swap dup f* f+ fsqrt ;
: z* { d c b a }
  a c f* b d f* f- a d f* c b f* f+
;
: _z*
  3 pick 2 pick f* 3 pick 2 pick f* f- 4 pick 2 pick f* 3 pick 5 pick f* f+
  2nip 2nip
;

: abs
  dup
  0< if
    negate
  then
;

: fak ( n -- n! ) 1+ 1 swap 1 ?do i * loop ;
: fakr ( n -- n! ) { x } x 0= if 1 else x x 1- fakr * then ;
: fakr2 ( n -- n! ) dup 0= if drop 1 else dup 1- fakr2 * then ;

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
