;;; globstar recognizes 3 types of captures to build scopes:
;;;
;;; - @local.scope: creates a new scope at the captured node
;;; - @local.definition: creates a definition in the current scope
;;; - @local.reference: creates reference to an existing definition, if any. If no definition
;;;   exists, a dangling reference is created.

;;; scopes
(type_alias_declaration) @local.scope
(type_declaration) @local.scope
(type_annotation) @local.scope
(port_annotation) @local.scope
(infix_declaration) @local.scope
(let_in_expr) @local.scope

;;; defs
(function_declaration_left (lower_pattern (lower_case_identifier)) @local.definition)
(function_declaration_left (lower_case_identifier) @local.definition)

;;; refs
(value_expr(value_qid(upper_case_identifier)) @local.reference)
(value_expr(value_qid(lower_case_identifier)) @local.reference)
(type_ref (upper_case_qid) @local.reference)

