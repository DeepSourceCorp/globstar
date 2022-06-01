;; scopes
(document) @local.scope
(block_node) @local.scope

;; defs
(block_mapping_pair
  key: (_) @local.definition
  value: (_))

(anchor (anchor_name) @global.definition)
(alias (alias_name) @global.reference)
