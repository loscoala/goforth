#!../goforth -file

\ : bounded? zdup zabs 2.0 f<= ;
: bounded? zdup dup f* swap dup f* f+ 4.0 f<= ;

: X-POS 0.43 ;
: Y-POS 0.33 fnegate ;
\ : Y-POS 0.2166393884377127 fnegate ;

: mb-kernel
  0 { i n ci c }
  0. 0. \ z0
  begin
    bounded?
    n i >
    and
  while
    \ zn+1 = (zn * zn) + c
    zdup z* c ci z+
    i 1+ to i
  repeat
  zdrop
  i
;

: X-SIZE 128 ;
: Y-SIZE 64 ;

: F-X-SIZE 128. ;
: F-Y-SIZE 64. ;

: scale-factor 0.9 ;
: scale-y i->f F-Y-SIZE f/ scale-factor f* Y-POS f+ ;
: scale-x i->f F-X-SIZE f/ F-X-SIZE F-Y-SIZE f/ f- scale-factor f* X-POS f+ ;

: mb-init !" 0  ====+++++++++********#########%%%%%%%%@@@@@@@@@@@@" ;
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
