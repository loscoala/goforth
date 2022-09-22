#!goforth -file

\ : bounded? zdup zabs 2.0 f<= ;
: bounded? zdup dup f* swap dup f* f+ 4.0 f<= ;

: X-POS 3.0 ;
: Y-POS 1.0 fnegate ;
\ : Y-POS 0.2166393884377127 fnegate ;

: mb-kernel
  0 { _i n ci c }
  0. 0. \ z0
  begin
    bounded?
    n _i >
    and
  while
    \ zn+1 = (zn * zn) + c
    zdup z* c ci z+
    _i 1+ to _i
  repeat
  zdrop
  _i
;

: X-SIZE 128 ;
: Y-SIZE 64 ;

: F-X-SIZE 128. ;
: F-Y-SIZE 64. ;

: scale-factor 2.3 ;
: scale-y i>f F-Y-SIZE f/ scale-factor f* Y-POS f+ ;
: scale-x i>f F-X-SIZE f/ F-X-SIZE F-Y-SIZE f/ f- scale-factor f* X-POS f+ ;

: mb-init
  memsize 55 < if 55 allocate then
  0 s"  ====+++++++++********#########%%%%%%%%@@@@@@@@@@@@"
;

: mb-draw 20 / 1+ @ emit ;

: mb
  0 0 { ci c }

  Y-SIZE 0 do
    i scale-y to ci
    X-SIZE 0 do
      i scale-x to c
      c ci 1000 mb-kernel
      mb-draw
    loop
    cr
  loop
;

: main
  mb-init
  mb
;
