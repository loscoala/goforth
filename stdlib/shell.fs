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
  @file@ [ a" bat #file#" system ] alloc
;

: inline vim
  @numArgs 0 @push @> @if
    @file@ [ a" vim #file#" system ] alloc
  @else
    [ a" vim" system ] alloc
  @then
;