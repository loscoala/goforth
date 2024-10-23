use llist.fs

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

