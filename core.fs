
: ? @ . ;
: ?? memsize 0 mem ;
: mem ?do i @ ?dup if i . colon . space then loop ;

: map swap dup @ over + 1+ swap 1+ ?do i @ over exec i ! loop drop ;
: each swap dup @ over + 1+ swap 1+ ?do i @ over exec loop drop ;
: bi4 ( n a b -- na nb ) { b a } dup a exec swap b exec ;

\ 1 [ dup 10 < ] [ ." Hello" 1+ ] while*
: while* { w b } begin b exec while w exec repeat drop ;

\ 10 0 [ . ] for
: for ( u l b -- ) { b } ?do i b exec loop ;

: s2i ( adr -- end begin ) [ @ over + 1+ ] [ 1+ ] bi4 ;
: each3 swap s2i [ @ over exec ] for drop ;
: map2 swap s2i [ { x } x @ over exec x ! ] for drop ;

\ find indices of bl in a string
\ variable code
\ variable my
\ 20 to code
\ my stack
\ code s" a b bs dsd s"
\ my code tok
: tok { adr stk } adr @ 0> if adr @ 1 do adr i + @ bl = if i 1 - stk push then loop then ;

: each1 \ a f --
  swap \ f a -
  dup  \ f a a --
  @    \ f a l --
  over \ f a l a --
  +    \ f a l+a --
  1+   \ f a l+a+1 --
  >r   \ f a -- end
  1+   \ f it -- end
  begin
    dup \ f it it -- end
    r@  \ f it it end -- end
    <   \ f it b -- end
  while
    dup  \ f it it -- end
    @    \ f it value -- end
    2 pick \ f it value f -- end
    exec   \ f it -- end
    1+
  repeat
  r>
  drop
  drop
  drop
;

: each2   \ f a --
  { f a } \     --
  a @     \ l   --
  a + 1+  \ l+a+1 --
  >r      \ -- end
  a 1+    \ it -- end
  begin
    dup
    r@
    <
  while
    dup
    @
    f
    exec
    1+
  repeat
  r>
  2drop
;

: map1 swap dup @ over + 1+ >r 1+ begin dup r@ < while dup @ 2 pick exec over ! 1+ repeat r> drop drop drop ;

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

: if* ( n a b -- ) { b a } if a exec else b exec then ;
\ : if* rot if drop else nip then exec ;
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
: spin -rot swap ;
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
: read ( name-addr buffer-size -- ) 8 sys ;
: debug ( bool -- ) 9 sys ;
: allocate ( size -- ) 10 sys ;
: memsize ( -- size ) 11 sys ;
: allot ( size -- adr ) memsize >r r@ + allocate r> ;
: alloc ( n a -- ) memsize >r swap allot swap exec drop r> allocate ;
: compare ( str1 str2 -- bool ) 12 sys ;
: shell ( str -- ) 13 sys ;
: system ( str -- ) 14 sys ;
: file ( str -- bool ) 15 sys ;
: argc ( -- n ) 16 sys ;
: argv ( addr n -- ) 17 sys ;

: >> over swap exec dup ;
: sh >> shell ;
: file-exist? >> file ;

\ : >> ( n a b -- n ) { b a } dup dup a exec b exec ;
\ : sh &shell >> ;
\ : file-exist? &file >> ;
: ls [ s" ls -lhA --color=always" ] sh ;

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
: _z*
  3 pick 2 pick f* 3 pick 2 pick f* f- 4 pick 2 pick f* 3 pick 5 pick f* f+
  2nip 2nip
;

: 2! tuck ! cell+ ! ;
: 2@ dup cell+ @ swap @ ;
: ++ 1 swap +! ;
: -- 1 negate swap +! ;
: @+ dup cell+ swap @ ;

\ Stack operations
: stack empty ;
: empty 0 swap ! ;
: push dup ++ dup @ + ! ;
: pop dup dup @ + @ swap -- ;
: .stack @+ ?dup if 0 ?do dup ? cell+ cr loop then drop ;

\ 100 [
\   dup s" He"
\   100 [
\     dup s" llo" 2dup append
\   ] alloc
\   dup .s6
\ ] alloc
: append ( t f -- ) s2i ?do dup i @ swap push loop drop ;
: append2 [ { v } over { adr } v adr push ] each3 drop ;

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
  2>r  \ 0 n1 n2 ... -- adr_end adr_i
  begin
    dup 0>
  while
    r> 1+ >r r@ !
  repeat
  drop
  2r>  \ adr_end adr_i
  over \ adr_end adr_i adr_end
  -    \ adr_end n
  swap \ n adr_end
  !
;

: !a ( 0 nx .. na n -- adr )
  allot
  dup
  >r
  !s
  r>
;

: .s2 { pos }
  pos @
  pos 1+ to pos
  0 ?do
    pos dup .c 1+ to pos
  loop
;

: .s3 [ emit ] each ;
: .s4 [ emit ] each1 ;
: .s5 [ emit ] each2 ;
: .s6 [ emit ] each3 ;

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

: showAlphabetPerChar2
  26 0 [ toChar space colon space 26 0 [ toChar ] for cr ] for
;

: integers 1+ 1 ?do i . space loop ;

