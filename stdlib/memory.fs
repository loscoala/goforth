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

: $ depth begin dup 0> while dup pick . space 1- repeat drop ;
: empty begin depth 0> while drop repeat ;

\ some custom words

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
