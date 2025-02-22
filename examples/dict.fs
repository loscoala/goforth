use llist.fs

( *
  * A dict is a list of pointers to kv structs to key, value of strings
  * )

: class dict extends list ;
: class kv 1 key 1 value ;

: kv:print ( self -- )
  dup kv:key @ .s ."  : " kv:value @ .s
;

: dict:append ( k v self -- )
  kv:new { kv self v k }
  k kv kv:key !
  v kv kv:value !
  kv self list:append
;

\ prints dict[string][string]
: dict:print ( self -- )
  [ kv:print cr ] swap dict:each
;
