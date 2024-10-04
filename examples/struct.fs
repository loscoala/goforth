\ To define a struct use:
\ struct <prefix:name> <size> <name>

struct foo 1 a 1 b 1 c

\ This generates the following words
\ foo:sizeof ( adr )
\ foo:a ( adr )
\ foo:b ( adr )
\ foo:c ( adr )
\ foo:[] ( index adr )

struct abc 25 a 12 b 10 c
struct xyz 1 a 2 b 5 c

