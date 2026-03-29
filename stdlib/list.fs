: class list
  len
  head
  tail
;

: class node
  data
  next
;

: node:print { self }
  self node:data ? ."  -> " self node:next ?
;

: list:append { self }
  node:new { node }
  node node:setData
  self list:getLen 0= if
    node self list:setHead
    node self list:setTail
  else
    node self list:getTail node:setNext
    node self list:setTail
  then
  self list:len ++
;

: list:each { self block }
  self list:getHead { current }
  begin
    current 0<>
  while
    current node:getData block exec
    current node:getNext to current
  repeat
;

: list:print ( self -- )
  \ [ node:print cr ] swap list:each
  [ . cr ] swap list:each
;

: list:map { self block }
  self list:getHead { current }
  begin
    current 0<>
  while
    current node:data dup @ block exec swap !
    current node:getNext to current
  repeat
;

: list:test
  list:new { l }
  98 l list:append
  97 l list:append
  96 l list:append
  l list:print
;
