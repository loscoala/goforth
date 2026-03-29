: class csv
  it     \ input string sviter
  line
  column
  state
  osv    \ output string
;

: csv:fromSV { sv }
  csv:new { self }
  sv sv:iter self csv:setIt
  sv:new { osv }
  osv self csv:setOsv
  0 osv sv:setLen
  255 allot osv sv:setData
  self
;

: csv:setOutLen ( self len )
  csv:getOsv sv:setLen
;

: csv:getOutLen ( self )
  csv:getOsv sv:getLen
;

: csv:appendCharToOut { self c }
  self csv:getOsv { sv }
  c sv sv:getData sv sv:getLen + !
  sv sv:len ++
;

: csv:print ( self )
  csv:getOsv sv:print
;

: csv:consume { self c }
  self csv:getState case
    0 of
      c case
        34 of 1 self csv:setState endof
        59 of self csv:column ++ endof
        10 of
          self csv:line ++
          0 self csv:setColumn
        endof
      endcase
      drop
    endof
    1 of
      c case
        34 of 0 self csv:setState endof
        c self csv:appendCharToOut
      endcase
      drop
    endof
  endcase
  drop
;

: csv:hasNext
  csv:getIt { it }
  it sviter:next
  it sviter:back
;

: csv:next { self }
  self csv:getIt { it }

  0 self csv:setOutLen

  begin
    it sviter:next
  while
    it sviter:get self csv:consume

    self csv:getState 0 =
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
    ." line: " self csv:getLine . ."  col: " self csv:getColumn . ."  "
    self csv:print cr
  ]
  a" \"ABC\";\"DEF\";\"GHI\";\"Udo\";\"Armin\";\"123\"\n\"ABC1\";\"DEF1\";\"GHI1\";\"Udo1\";\"Armin1\";\"1231\""
  csv:fromSV csv:parseCSV
;