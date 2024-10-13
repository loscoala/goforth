: class string 1 len 1 items ;

: string:each ( block self -- )
  { self block }
  self string:len @ 0 ?do
    self string:items i + @ block exec
  loop
;

: string:print ( self -- )
  [ emit ] swap string:each
;

