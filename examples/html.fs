#!goforth -file

: document ." <!doctype html>" cr ;

: tag ." <" exec ." >" ;
: /tag ." </" exec ." >" ;

: enclose dup tag swap exec /tag ;

: html [ ." html lang=\"de-DE\"" ] tag exec [ ." html" ] /tag ;
: head [ ." head" ] enclose ;
: body [ ." body" ] enclose ;

: meta [ ." meta " exec ] tag ;
: title [ ." title" ] enclose ;
: p [ ." p" ] enclose ;
: h1 [ ." h1" ] enclose ;

: main
  document
  [
    [
      [ ." charset=\"utf-8\"" ] meta
      [ ." Example Page" ] title
    ] head
    [
      [ ." Example Page" ] h1
      [ ." Hello World!" ] p
    ] body
  ] html
;
