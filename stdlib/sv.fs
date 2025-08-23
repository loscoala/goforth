: class sv
  1 len
  1 data
;

: sv:fromS ( 0 c b a N -- adr )
  sv:new { self len }
  len self sv:len !
  len allot self sv:data !
  self sv:data @ { ptr }
  begin
    dup 0>
  while
    ptr !
    ptr 1+ to ptr
  repeat
  drop
  self
;

: sv:print { self }
  self sv:data @ { ptr }
  self sv:len @
  begin
    dup 0>
  while
    ptr @ emit
    ptr 1+ to ptr
    1-
  repeat
  drop
;

: inline sv:_toS { _self }
  _self sv:data @ _self sv:len @ 1- +
  _self sv:data @
  { _base _ptr }
  begin
    _ptr _base >=
  while
    _ptr @
    _ptr 1- to _ptr
  repeat
  _self sv:len @
  done
  done
;

: sv:toS { self }
  0
  self sv:_toS
;

: sv:append { self other }
  other sv:toS drop self sv:_toS drop
  other sv:len @
  self sv:len @
  +
;
