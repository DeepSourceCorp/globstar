;;; shell injections in ansible can occur in one of two modules
;;; - ansible.builtins.command
;;; - ansible.builtins.shell

((block_mapping_pair
   key: (_) @key
   value:
   (flow_node
     (plain_scalar  ; TODO: strip quotes from quoted scalars and parse as bash
       (string_scalar) @injection.content)))
 (#match? @key "^(ansible\\.builtin\\.)?(command|shell)")
 (#set! injection.language "bash"))

