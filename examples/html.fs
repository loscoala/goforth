#!goforth -file

: document ." <!doctype html>" cr ;
: tag ." <" exec ." >" ;
: /tag ." </" exec ." >" ;
: html [ ." html lang=\"de-DE\"" ] tag ;
: /html [ ." html" ] /tag ;
: head ." <head>" ;
: /head ." </head>" ;
: meta [ ." meta charset=\"utf-8\"" ] tag ;
: title ." <title>" ;
: /title ." </title>" ;
: body ." <body>" ;
: /body ." </body>" ;
: p ." <p>" ;
: /p ." </p>" ;
: h1 ." <h1>" ;
: /h1 ." </h1>" ;

: main
  document
  html
    head
      meta
      title ." Example Page" /title
    /head
    body
      h1 ." Example Page h1" /h1
      p ." Hello World!" /p
    /body
  /html
;
