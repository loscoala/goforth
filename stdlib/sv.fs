: class sv
  1 len
  1 data
;

: sv:fromS ( 0 c b a N -- adr )
  sv:new { self len }
  len self sv:len !
  len allot self sv:data !
  self sv:data @ { ptr }
  [ dup 0> ]
  [
    ptr !
    ptr 1+ to ptr
  ]
  while!
  self
;

: sv:print { self }
  self sv:data @ { ptr }
  self sv:len @
  [ dup 0> ]
  [
    ptr @ emit
    ptr 1+ to ptr
    1-
  ]
  while!
;

: sv:each { self block }
  [
    self sv:iter
    [ dup sviter:next ]
    [ dup sviter:get block exec ]
    while!
  ] alloc
;

: inline sv:_toS @1@
  #1# sv:data @ #1# sv:len @ 1- +
  #1# sv:data @
  { base ptr }
    begin
      ptr base >=
    while
      ptr @
      ptr 1- to ptr
    repeat
  done
;

: sv:toS { self }
  0
  self sv:_toS
  self sv:len @
;

: sv:append { self other }
  0
  other sv:_toS
  self sv:_toS
  other sv:len @
  self sv:len @
  +
;

\ ------------ Iterator ----------------

: class sviter
  1 sv
  1 len
  1 index
;

: sv:iter { self }
  sviter:new { it }
  self it sviter:sv !
  self sv:len @ it sviter:len !
  -1 it sviter:index !
  it
;

: sviter:next { self }
  self sviter:index @ self sviter:len @ 1- < if
    self sviter:index ++
    true
  else
    false
  then
;

: sviter:back { self }
  self sviter:index --
;

: sviter:get { self }
  self sviter:sv @ sv:data @ self sviter:index @ + @
;
