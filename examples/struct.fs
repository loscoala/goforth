\ To define a struct use:
\ struct <name> <size> <name> ...

struct foo 1 a 1 b 1 c

\ This generates the following words
\ foo:allot ( n -- adr )
\ foo:sizeof ( -- n )
\ foo:a ( adr -- adr2 )
\ foo:b ( adr -- adr2 )
\ foo:c ( adr -- adr2 )
\ foo:[] ( index adr -- adr2 )

struct abc 25 a 12 b 10 c
struct xyz 1 a 2 b 5 c

