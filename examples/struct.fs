\ To define a class use:
\ class <name> <size> <name> ...

: class foo 1 a 1 b 1 c ;

\ This generates the following words
\ foo:allot ( n -- adr )
\ foo:sizeof ( -- n )
\ foo:a ( adr -- adr2 )
\ foo:b ( adr -- adr2 )
\ foo:c ( adr -- adr2 )
\ foo:[] ( index adr -- adr2 )

: class moo extends foo ;
: class hoo extends foo 1 d 1 e 1 f ;

: class abc 25 a 12 b 10 c ;
: class xyz 1 a 2 b 5 c ;

