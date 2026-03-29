: class csv
  1 it     \ input string sviter
  1 line
  1 column
  1 state
  1 osv    \ output string
;

: csv:fromSV { sv }
  csv:new { self }
  sv sv:iter self csv:it !
  sv:new { osv }
  osv self csv:osv !
  0 osv sv:len !
  255 allot osv sv:data !
  self
;

: csv:setOutLen ( self len )
  csv:getOutSV sv:len !
;

: csv:getOutLen ( self )
  csv:getOutSV sv:len @
;

: csv:appendCharToOut { self c }
  self csv:getOutSV { sv }
  c sv sv:data @ sv sv:len @ + !
  sv sv:len ++
;

: csv:getOutSV ( self )
  csv:osv @
;

: csv:print ( self )
  csv:getOutSV sv:print
;

: csv:consume { self c }
  self csv:state @ case
    0 of
      c case
        34 of 1 self csv:state ! endof
        59 of self csv:column ++ endof
        10 of
          self csv:line ++
          0 self csv:column !
        endof
      endcase
      drop
    endof
    1 of
      c case
        34 of 0 self csv:state ! endof
        c self csv:appendCharToOut
      endcase
      drop
    endof
  endcase
  drop
;

: csv:hasNext
  csv:it @ { it }
  it sviter:next
  it sviter:back
;

: csv:next { self }
  self csv:it @ { it }

  0 self csv:setOutLen

  begin
    it sviter:next
  while
    it sviter:get self csv:consume

    self csv:state @ 0 =
    self csv:getOutLen 0 >
    and if
      leave
    then
  repeat
;

: csv:parseCSV { self block }
  begin
    self csv:hasNext
  while
    self csv:next
    self block exec
  repeat
;

: csv:parseFile ( sv block )
  readfile sv:fromS csv:fromSV csv:parseCSV
;

: csv:test
  [
    { self }
    ." line: " self csv:line @ . ."  col: " self csv:column @ . ."  "
    self csv:print cr
  ]
  a" \"ABC\";\"DEF\";\"GHI\";\"Udo\";\"Armin\";\"123\"\n\"ABC1\";\"DEF1\";\"GHI1\";\"Udo1\";\"Armin1\";\"1231\""
  csv:fromSV csv:parseCSV
;