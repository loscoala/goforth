: class list
  1 len
  1 head
  1 tail
;

: class node
  1 data
  1 next
;

: node:print { self }
  self node:data ? ."  -> " self node:next ?
;

: list:append { self }
  1 node:allot { node }

  node node:data !
  0 node node:next !
  
  self list:len @ 0= if
    node self list:head !
    node self list:tail !
  else
    node self list:tail @ node:next !
    node self list:tail !
  then

  self list:len ++
;

: list:each { self block } 
  self list:head @ { current }
  begin
    current 0<>
  while
    current block exec
    current node:next @ to current
  repeat
;

: list:print ( self -- )
  [ node:print cr ] swap list:each
;

: list:test
  1 list:allot drop
  98 0 list:append
  97 0 list:append
  96 0 list:append
  0 list:print
;

