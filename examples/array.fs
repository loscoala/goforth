: class array
  1 len
  1 items
  1 capacity
;

: array:append { self item }
  self array:len @ self array:capacity @ >= if
    self array:capacity @ 0 = if
      256 self array:capacity !
    else
      self array:capacity @ 2* self array:capacity !
    then
    self array:capacity @ allot self array:items !
  then
  \ push value
  item self array:len @ self array:items @ th !
  \ inc len
  self array:len ++
;

: array:clear
  0 swap array:len !
;

: array:print
  [ . cr ] swap array:each
;

: array:each { self block }
  self array:len @ 0 ?do
    self array:items @ i + @ block exec
  loop
;

: array:at { self block index }
  self array:items @ index + @ block exec
;

: array:map { self block }
  self array:len @ 0 ?do
    self array:items @ i + dup @ block exec swap !
  loop
;

