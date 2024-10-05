struct stack 1 len 200 items

: stack:push { self value }
  \ push value
  value self stack:len @ self stack:items th !
  \ inc len
  self stack:len @ 1+ self stack:len !
;

: stack:pop { self }
  \ dec len
  self stack:len @ 1- self stack:len !
  \ pop value
  self stack:len @ self stack:items th @
;
