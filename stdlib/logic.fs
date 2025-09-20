: inline if+ @f@ @t@ if #t# else #f# then ;
: if* ( n a b -- ) { b a } if a exec else b exec then ;

: inline times+ @a@ 0 ?do #a# loop ;
: times swap 0 ?do dup exec loop drop ;

: inline when+ @a@ if #a# then ;
: when swap if exec else drop then ;

: inline bi+ @b@ @a@ dup #a# swap #b# ;
: bi ( n a b -- na nb ) { b a } dup a exec swap b exec ;

: tri { a b c } dup c exec over b exec rot a exec ;
: inline tri+ @a@ @b@ @c@ dup #c# over #b# rot #a# ;

: dip ( x quot -- x ) swap { x } exec x ;
: keep ( ..a x quot -- ..b x ) over { x } exec x ;

: inline curry+ @1@ @2@ [ #2# #1# ] ;

\ 1 [ dup 10 < ] [ ." Hello" 1+ ] while*
: while* { w b } begin b exec while w exec repeat drop ;
: inline while+ @w@ @b@ begin #b# while #w# repeat drop ;

\ 10 0 [ . ] for
: for ( u l b -- ) { b } ?do i b exec loop ;
: inline for+ @b@ ?do i #b# loop ;
