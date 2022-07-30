
: ? @ . ;
: ?? 0 1000 mem ;
: .?
  dup 31 >
  over
  127 <
  and
  if
    lb emit rb
  else
    drop
  then
;
: mem
  0 do
    dup i +
    dup @ 0<> if
      dup . colon ? ( @ dup . .? ) space
    else
      drop
    then
  loop
  drop
;
: map swap dup @ over + 1+ swap 1+ ?do i @ over exec i ! loop drop ;
: each swap dup @ over + 1+ swap 1+ ?do i @ over exec loop drop ;
\ : bi ( n a b -- na nb ) { _bi_b _bi_a } dup _bi_a exec swap _bi_b exec ;

(
: bi  \ n a b
  rot \ a b n
  dup \ a b n n
  rot \ a n n b
  exec \ a n nb
  -rot \ nb a n
  swap \ nb n a
  exec \ nb na
  swap \ na nb
;
)

: bi \ n a b
  2 pick \ n a b n
  swap \ n a n b
  exec \ n a nb
  -rot \ nb n a
  exec \ nb na
  swap \ na nb
;

: bi2 \ n a b
  >r  \ n a -- b
  over \ n a n -- b
  swap \ n n a -- b
  exec \ n na -- b
  swap \ na n -- b
  r>   \ na n b
  exec \ na nb
;

: bi3  \ n a b
  >r   \ n a -- b
  over \ n a n -- b
  r>   \ n a n b
  exec \ n a nb
  >r   \ n a -- nb
  exec \ na -- nb
  r>   \ na nb
;

\ : if* ( n a b -- ) { _if*_b _if*_a } if _if*_a exec else _if*_b exec then ;
: if* rot if drop else nip then exec ;
\ : ifb { _ifb_b _ifb_a } dup 0<> _ifb_a * swap 0= _ifb_b * + exec ;
\ : ifb2 { _ifb2_b _ifb2_a } [ 0<> _ifb2_a * ] [ 0= _ifb2_b * ] bi + exec ;

: times swap 0 ?do dup exec loop drop ;
\ : times swap begin dup 0> while swap dup exec swap 1- repeat 2drop ;
\ : times { _times2_a } 0 ?do _times2_a exec loop ;

: when swap if exec else drop then ;

\ : -rot rot rot ;
: nip swap drop ;
: 2drop drop drop ;
: 2nip 2swap 2drop ;
: tuck swap over ;
: negate 0 swap - ;
: fnegate 0. swap f- ;
: vswap
  { v1 v2 }
  v1 @ v2 @ v1 ! v2 !
;
: i>f 3 sys ;
: f>i 4 sys ;
: readfile ( dest-addr name-addr -- ) 5 sys ;
: readimage ( name-addr -- ) 6 sys ;
: writeimage ( name-addr -- ) 7 sys ;

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
: f>= f< not ;
: <= > not ;
: f<= f> not ;
: 0<> 0= not ;
: 0<= 0 <= ;
: 0>= 0 >= ;
: slow_mod begin 2dup >= while dup -rot - swap repeat drop ;
: mod 1 sys ;
: fsqrt 2 sys ;
: factor mod 0= ;
: even 2 factor ;

: min 2dup < if drop else nip then ;
: max 2dup > if drop else nip then ;

: false 0 ;
: true false not ;

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

: print begin dup 0> while emit repeat drop ;

: !s2 0 0 { n y x }
  x to y
  begin
    dup 0>
  while
    n 1+ to n
    y 1+ to y
    y !
  repeat
  n x !
  drop
;

: !s \ 0 n1 n2 ...  adr
  dup \ 0 n1 n2 ... adr adr --
  >r  \ 0 n1 n2 ... adr -- adr
  >r  \ 0 n1 n2 ... -- adr_end adr_i
  begin
    dup 0>
  while
    r> 1+ >r r@ !
  repeat
  drop
  r>   \ adr_i
  r>   \ adr_i adr_end
  dup  \ adr_i adr_end adr_end
  rot \ adr_end adr_end adr_i
  swap
  -   \ adr_end n
  swap
  !
;

: .s2 { pos }
  pos @
  pos 1+ to pos
  0 ?do
    pos dup .c 1+ to pos
  loop
;

: .s3 [ emit ] each ;

: +s dup ++ dup @ + ! ;
: -s -- ;

: abs
  dup
  0< if
    negate
  then
;

: fak ( n -- n! ) 1+ 1 swap 1 ?do i * loop ;

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

