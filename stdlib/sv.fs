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

: sv:toS { self }
  0
  self sv:data @ self sv:len @ 1- +
  self sv:data @
  { base ptr }
  begin
    ptr base >=
  while
    ptr @
    ptr 1- to ptr
  repeat
  self sv:len @
;
