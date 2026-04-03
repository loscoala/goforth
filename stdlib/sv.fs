: class sv
  len
  data
;

: sv:fromS ( 0 c b a N -- adr )
  sv:new { self len }
  len self sv:setLen
  len allot self sv:setData
  self sv:getData { ptr }
  [ dup 0> ]
  [
    ptr !
    ptr 1+ to ptr
  ]
  while!
  self
;

: sv:print { self }
  self sv:getData { ptr }
  self sv:getLen
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
  #1# sv:getData #1# sv:getLen 1- +
  #1# sv:getData
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
  self sv:getLen
;

: sv:append { self other }
  0
  other sv:_toS
  self sv:_toS
  other sv:getLen
  self sv:getLen
  +
;

\ ------------ Iterator ----------------

: class sviter
  sv
  len
  index
;

: sv:iter { self }
  sviter:new { it }
  self it sviter:setSv
  self sv:getLen it sviter:setLen
  -1 it sviter:setIndex
  it
;

: sviter:next { self }
  self sviter:hasNext if
    self sviter:index ++
    true
  else
    false
  then
;

: sviter:hasNext { self }
  self sviter:getIndex self sviter:getLen 1- <
;

: sviter:back { self }
  self sviter:index --
;

: sviter:get { self }
  self sviter:getSv sv:getData self sviter:getIndex + @
;
