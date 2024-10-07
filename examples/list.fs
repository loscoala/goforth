struct list 1 len 200 items

: list:append { self }
  \ push value
  self list:len @ self list:items th !
  \ inc len
  self list:len @ 1+ self list:len !
;

: list:clear
  0 swap list:len !
;

: list:print { self }
  self list:len @ 0 ?do
    self list:items i + ? cr
  loop
;

: list:each { self block }
  self list:len @ 0 ?do
    self list:items i + @ block exec
  loop
;

: list:at { self block index }
  self list:items index + @ block exec
;
