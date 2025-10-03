\ : inline if+ @f@ @t@ if #t# else #f# then ;
\ : if* ( n a b -- ) { b a } if a exec else b exec then ;
: inline if!
  @numArgs 1 @push @> @if
    @f@ @t@ if #t# else #f# then
  @else
    { b a } if a exec else b exec then
  @then
;

\ : inline times+ @a@ 0 ?do #a# loop ;
\ : times swap 0 ?do dup exec loop drop ;
: inline times!
  @numArgs 0 @push @> @if
    @a@ 0 ?do #a# loop
  @else
    swap 0 ?do dup exec loop drop
  @then
;

\ : inline when+ @a@ if #a# then ;
\ : when swap if exec else drop then ;
: inline when!
  @numArgs 0 @push @> @if
    @a@ if #a# then
  @else
    swap if exec else drop then
  @then
;

\ : inline bi+ @b@ @a@ dup #a# swap #b# ;
\ : bi ( n a b -- na nb ) { b a } dup a exec swap b exec ;
: inline bi!
  @numArgs 1 @push @> @if
    @b@ @a@ dup #a# swap #b#
  @else
    { b a } dup a exec swap b exec
  @then
;

\ : tri { a b c } dup c exec over b exec rot a exec ;
\ : inline tri+ @a@ @b@ @c@ dup #c# over #b# rot #a# ;
: inline tri!
  @numArgs 2 @push @> @if
    @a@ @b@ @c@ dup #c# over #b# rot #a#
  @else
    { a b c } dup c exec over b exec rot a exec
  @then
;

: dip ( x quot -- x ) swap { x } exec x ;
: keep ( ..a x quot -- ..b x ) over { x } exec x ;

: inline curry+ @1@ @2@ [ #2# #1# ] ;

\ 1 [ dup 10 < ] [ ." Hello" 1+ ] while*
\ : while* { w b } begin b exec while w exec repeat drop ;
\ : inline while+ @w@ @b@ begin #b# while #w# repeat drop ;
: inline while!
  @numArgs 1 @push @> @if
    @w@ @b@ begin #b# while #w# repeat drop
  @else
    { w b } begin b exec while w exec repeat drop
  @then
;

\ 10 0 [ . ] for
\ : for ( u l b -- ) { b } ?do i b exec loop ;
\ : inline for+ @b@ ?do i #b# loop ;
: inline for!
  @numArgs 0 @push @> @if
    @b@ ?do i #b# loop
  @else
    { b } ?do i b exec loop
  @then
;
