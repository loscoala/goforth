use llist.fs

\ fstr       [len][data ....]
\ stack str  g( ABC) -> 0 C B A 4 --

\ String_View
: class sv
  1 count
  1 data \ pointer
;

\ fstr -- *sv
: sv:from_fstr { fstr }
  sv:new { self }
  fstr @ self sv:count !
  fstr 1+ self sv:data !
  self
;

: sv:from_stack
  !a sv:from_fstr
;

: sv:print { self }
  self sv:count @ 0 do
    i self sv:data @ + @ emit
  loop
;

: sv:print2 { self }
  self sv:count @
  0
  begin
    dup 2 pick <
  while
    dup self sv:data @ + @ emit
    1+
  repeat
  2drop
;

\ StringBuilder
\ TODO: convert into stack str and String_View
: class string extends list
  1 slen
;

: string:length ( self -- )
  string:slen @
;

: string:append { self item }
  item @ self string:slen @ + self string:slen !
  item self list:append
;

: string:print ( self -- )
  &.s swap string:each
;

: string:test
  string:new { str }
  a" hello" str string:append
  a"  world" str string:append
  str string:print
;

