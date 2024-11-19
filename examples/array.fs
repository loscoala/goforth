: class array
  1 len
  1 items
  1 capacity
;

: array:INIT_CAP 256 ;

: memcpy { dest src n }
  begin
    n 0 >
  while
    src @ dest !
    src 1+ to src
    dest 1+ to dest
    n 1- to n
  repeat
;

: array:resize_items { self }
  self array:capacity @ allot { new_array }
  self array:len @ self array:items @ new_array memcpy
  new_array self array:items !
;

: array:append { self item }
  self array:len @ self array:capacity @ >= if
    self array:capacity @ 0 = if
      array:INIT_CAP self array:capacity !
    else
      self array:capacity @ 2* self array:capacity !
    then
    self array:resize_items
  then
  item self array:len @ self array:items @ th !
  self array:len ++
;

: array:append_many { self items num_items }
  self array:len @ num_items + self array:capacity @ > if
    self array:capacity @ 0 = if
      array:INIT_CAP self array:capacity !
    then
    begin
      self array:len @ num_items +
      self array:capacity @ >
    while
      self array:capacity @ 2* self array:capacity !
    repeat
    self array:resize_items
  then
  num_items items self array:items @ self array:len @ + memcpy
  self array:len @ num_items + self array:len !
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

