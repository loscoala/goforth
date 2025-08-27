: inline if+ @2@ if #1# else #2# then ;

: if* ( n a b -- ) { b a } if a exec else b exec then ;
: times swap 0 ?do dup exec loop drop ;
: when swap if exec else drop then ;
: bi ( n a b -- na nb ) { b a } dup a exec swap b exec ;

\ 1 [ dup 10 < ] [ ." Hello" 1+ ] while*
: while* { w b } begin b exec while w exec repeat drop ;

\ 10 0 [ . ] for
: for ( u l b -- ) { b } ?do i b exec loop ;

