#!goforth -file

\ This implementation is free of local variables and can be compiled into C

: bounded? zdup dup f* swap dup f* f+ 4.0 f<= ;

: X-POS 3.0 ;
: Y-POS 1.0 fnegate ;
\ : Y-POS 0.2166393884377127 fnegate ;

: _i 802 ;
: n 803 ;

: mb-kernel
  0 _i !
  0. 0. \ z0
  begin
    bounded?
    n @ _i @ >
    and
  while
    \ zn+1 = (zn * zn) + c
    zdup _z* c @ ci @ z+
    _i @ 1+ _i !
  repeat
  zdrop
  _i @
;

: X-SIZE 128 ;
: Y-SIZE 64 ;

: F-X-SIZE 128.0 ;
: F-Y-SIZE 64.0 ;

: scale-factor 2.3 ;
: scale-y i>f F-Y-SIZE f/ scale-factor f* Y-POS f+ ;
: scale-x i>f F-X-SIZE f/ F-X-SIZE F-Y-SIZE f/ f- scale-factor f* X-POS f+ ;

: mb-init
  memsize 1000 < if 1000 allocate then
  0 s"  ====+++++++++********#########%%%%%%%%@@@@@@@@@@@@"
;

: mb-draw 20 / 1+ @ emit ;

: ci 800 ;
: c 801 ;

: mb

  Y-SIZE 0 do
    i scale-y ci !
    X-SIZE 0 do
      i scale-x c !
      1000 n !
      mb-kernel
      mb-draw
    loop
    cr
  loop
;

: main
  mb-init
  mb
;
