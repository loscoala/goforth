: inline if+ @2@ if #1# else #2# then ;
: if* ( n a b -- ) { b a } if a exec else b exec then ;

: inline times+ @1@ 0 ?do #1# loop ;
: times swap 0 ?do dup exec loop drop ;

: inline when+ @2@ #2# if #1# then ;
: when swap if exec else drop then ;

: inline bi+ @2@ dup #2# swap #1# ;
: bi ( n a b -- na nb ) { b a } dup a exec swap b exec ;

\ 1 [ dup 10 < ] [ ." Hello" 1+ ] while*
: while* { w b } begin b exec while w exec repeat drop ;
: inline while+ @2@ begin #2# while #1# repeat drop ;

\ 10 0 [ . ] for
: for ( u l b -- ) { b } ?do i b exec loop ;
: inline for+ @1@ ?do i #1# loop ;
