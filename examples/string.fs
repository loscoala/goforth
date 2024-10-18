use llist.fs

: class string extends list ;

: string:print ( self -- )
  &.s swap string:each
;

: string:test
  string:new { str }
  a" hello" str string:append
  a"  world" str string:append
  str string:print
;

