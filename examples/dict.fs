use list.fs

( *
  * A dict is a list of pointer to kv structs
  * )

struct dict extends list
struct kv 1 key 1 value

: dict:append ( k v self -- )
  1 kv:allot { kv self v k }
  k kv kv:key !
  v kv kv:value !
  kv self list:append
;

