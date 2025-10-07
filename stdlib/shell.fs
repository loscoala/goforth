( ***************************
  macro ends at | word
  : test
    | ls -lhA $$
    | htop $$
    | btop $$
    | du -sh . $$
  ;
  *************************** )
: inline $$
  @push @dup | @push @= @if
    @drop [ @$ ] @1@ [ a" #1#" shell ] alloc
  @else
    $$
  @then
;

: btop
  | btop $$
;

: htop
  | htop $$
;

: ls
  | eza -lhA --color=always $$
;

: build
  | go build -C cmd/goforth $$
;

: inline bat
  @file@ [ a" bat #file#" shell ] alloc
;

: inline vim
  @numArgs 0 @push @> @if
    @file@ [ a" vim #file#" shell ] alloc
  @else
    [ a" vim" shell ] alloc
  @then
;