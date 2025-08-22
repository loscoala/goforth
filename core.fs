
: ? @ . ;
: ?? memsize 0 mem ;
: mem ?do i @ ?dup if i . colon . space then loop ;

: bi ( n a b -- na nb ) { b a } dup a exec swap b exec ;

\ 1 [ dup 10 < ] [ ." Hello" 1+ ] while*
: while* { w b } begin b exec while w exec repeat drop ;

\ 10 0 [ . ] for
: for ( u l b -- ) { b } ?do i b exec loop ;

: memset { dest c n }
  begin
    n 0 >
  while
    c dest !
    dest 1+ to dest
    n 1- to n
  repeat
;

: memcpy { dest src n }
  begin
    n 0 >
  while
    src @ dest !
    src 1+ to src
    dest 1+ to dest
    n 1- to n
  repeat
;

: if* ( n a b -- ) { b a } if a exec else b exec then ;
: times swap 0 ?do dup exec loop drop ;
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
: readfile ( name-addr -- ) 5 sys ;
: readimage ( name-addr -- ) 6 sys ;
: writeimage ( name-addr -- ) 7 sys ;
: read ( buffer-size -- ) 8 sys ;
: debug ( bool -- ) 9 sys ;
: allocate ( size -- ) 10 sys ;
: memsize ( -- size ) 11 sys ;
variable here
: allot ( size -- adr )
  here >r
  here + to here
  here memsize > if
    here 2* allocate
  then
  r>
;
: alloc ( block -- ) here swap exec to here ;
: compare ( str1 str2 -- bool ) 12 sys ;
: shell ( str -- ) 13 sys ;
: system ( str -- ) 14 sys ;
: file ( str -- bool ) 15 sys ;
: argc ( -- n ) 16 sys ;
: argv ( addr n -- ) 17 sys ;

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
: print begin dup 0> while emit repeat drop ;

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

\ ---------------- STRING ---------------------

: class sv
  1 len
  1 data
;

: sv:fromS ( 0 c b a N -- adr )
  sv:new { self len }
  len self sv:len !
  len allot self sv:data !
  self sv:data @ { ptr }
  begin
    dup 0>
  while
    ptr !
    ptr 1+ to ptr
  repeat
  drop
  self
;

: sv:print { self }
  self sv:data @ { ptr }
  self sv:len @
  begin
    dup 0>
  while
    ptr @ emit
    ptr 1+ to ptr
    1-
  repeat
  drop
;

: sv:toS { self }
  0
  self sv:data @ self sv:len @ 1- +
  self sv:data @
  { base ptr }
  begin
    ptr base >=
  while
    ptr @
    ptr 1- to ptr
  repeat
  self sv:len @
;

