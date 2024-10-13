use list.fs

: class stack extends list ;

: stack:push { self }
  \ push value
  self stack:len @ self stack:items th !
  \ inc len
  self stack:len @ 1+ self stack:len !
;

: stack:pop { self }
  \ dec len
  self stack:len @ 1- self stack:len !
  \ pop value
  self stack:len @ self stack:items th @
;

