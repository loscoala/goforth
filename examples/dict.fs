use list.fs

( *
  * A dict is a list of pointers to kv structs to key, value of strings
  * )

struct dict extends list
struct kv 1 key 1 value

: kv:print ( self -- )
  dup kv:key @ .s ."  : " kv:value @ .s
;

: dict:append ( k v self -- )
  1 kv:allot { kv self v k }
  k kv kv:key !
  v kv kv:value !
  kv self list:append
;

\ prints dict[string][string]
: dict:print ( self -- )
  [ kv:print cr ] swap dict:each
;
