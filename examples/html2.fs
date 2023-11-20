template test html.tfs
variable title

: run
  100 [
    dup s" Hello World"
    dup to title
    test
  ] alloc
;
