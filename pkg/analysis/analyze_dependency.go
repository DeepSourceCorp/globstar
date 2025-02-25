package analysis

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// Attempt at creating a generic dependency extracting query for all the supported languages.
func GetDependencyQuery(lang Language) (string, error) {
    var queryStr string

    switch lang {
    case LangPy:
        queryStr = `
            (import_statement
                name: (dotted_name) @import)
            (import_from_statement
                module_name: (dotted_name) @from)
            (import_from_statement
                module_name: (relative_import) @relative_import)
            (call
                function: (identifier) @func (#match? @func "^(__import__|importlib\\.import_module)$")
                arguments: (argument_list (string) @dynamic_import))
        `
    case LangJs, LangTs, LangTsx:
        queryStr = `
            (import_statement
                source: (string) @import)
            (import_specifier
                name: (identifier) @import_name)
            (call_expression
                function: [
                    (identifier) @require (#eq? @require "require")
                    (member_expression (identifier) @require (#eq? @require "require"))
                ]
                arguments: (arguments (string) @path))
            (call_expression
                function: (identifier) @import (#eq? @import "import")
                arguments: (arguments (string) @dynamic_import))
            (export_statement
                source: (string) @export_from)
        `
    case LangJava:
        queryStr = `
            (import_declaration
                (scoped_identifier) @import)
            (package_declaration
                (scoped_identifier) @package)
            (method_invocation
                object: (identifier) @class (#eq? @class "Class")
                name: (identifier) @method (#eq? @method "forName")
                arguments: (argument_list (string_literal) @dynamic_import))
        `
    case LangRuby:
        queryStr = `
            (call
                method: (identifier) @method (#match? @method "^(require|require_relative|load|autoload|require_dependency|import|include|extend)$")
                arguments: (argument_list (string) @path))
            (class (constant) @class)
            (module (constant) @module)
            (assignment
                left: (constant) @constant
                right: (call
                        method: (identifier) @method (#eq? @method "require")
                        arguments: (argument_list (string) @dynamic_require)))
        `
    case LangRust:
        queryStr = `
            (use_declaration
                path: (path) @import)
            (extern_crate_declaration
                name: (identifier) @crate)
            (macro_invocation
                macro: (identifier) @include (#match? @include "^(include|include_bytes|include_str)$")
                arguments: (token_tree (string_literal) @path))
        `
    case LangSql:
        // SQL doesn't typically have imports, but we can look for includes
        queryStr = `
            (include_statement
                file_name: (_) @include)
            (\@include (_) @include_directive)
            (copy_statement source: (_) @copy_source)
            (execute_statement command: (_) @execute_command)
        `
    case LangKotlin:
        queryStr = `
            (import_header
                identifier: [
                    (identifier) @import
                    (navigation_expression) @qualified_import
                ])
            (package_header
                identifier: [
                    (identifier) @package
                    (navigation_expression) @qualified_package
                ])
            (call_expression
                expression: (navigation_expression) @dynamic_import (#match? @dynamic_import ".*[Cc]lass\\.forName"))
        `
    case LangCss:
        queryStr = `
            (import_statement
                url: (string_value) @import)
            (at_rule
                name: (at_keyword) @rule (#eq? @rule "@import")
                value: (prelude [
                    (string_value) @path
                    (url) @url_path
                ]))
            (at_rule
                name: (at_keyword) @rule (#eq? @rule "@use")
                value: (prelude (string_value) @use_path))
            (at_rule
                name: (at_keyword) @rule (#eq? @rule "@forward")
                value: (prelude (string_value) @forward_path))
        `
    case LangOCaml:
        queryStr = `
            (open_statement
                module_path: (module_path) @open)
            (include_statement
                module_expression: [
                    (module_path) @include
                    (module_expression) @complex_include
                ])
            (external_declaration
                primitive: (string_literal) @external)
        `
    case LangLua:
        queryStr = `
            (function_call
                name: (identifier) @require (#eq? @require "require")
                arguments: (arguments (string) @path))
            (function_call
                name: (identifier) @load (#match? @load "^(load|loadfile|dofile)$")
                arguments: (arguments (string) @load_path))
            (function_call
                name: (method_index
                    table: (identifier) @package (#eq? @package "package")
                    method: (identifier) @method (#eq? @method "loadlib"))
                arguments: (arguments (string) @lib_path))
        `
    case LangDockerfile:
        queryStr = `
            (from_instruction
                image: (image_spec 
                    (image_name) @from
                    (image_tag)? @tag
                    (image_digest)? @digest))
            (copy_instruction
                source: (path) @copy_source)
            (run_instruction
                command: (shell_command) @run_command (#match? @run_command "^(npm|yarn|pip|go|apt|apk|dnf|yum)"))
        `
    case LangBash:
        queryStr = `
            (command
                name: (command_name) @command (#match? @command "^(source|\\.|sh|bash|zsh|exec)$")
                argument: (word) @source_path)
            (command
                name: (command_name) @import_cmd (#match? @import_cmd "^(import|include|require|load)$")
                argument: (word) @import_path)
            (command
                name: (command_name) @pkg_cmd (#match? @pkg_cmd "^(apt-get|apt|yum|dnf|pacman|brew|npm|pip|gem)$")
                argument: (word) @install_arg (#eq? @install_arg "install")
                argument: (word) @package)
        `
    case LangCsharp:
        queryStr = `
            (using_directive
                name: (qualified_name) @using)
            (namespace_declaration
                name: (qualified_name) @namespace)
            (attribute
                name: (qualified_name) @attribute)
            (invocation_expression
                expression: (member_access_expression) @dynamic_import (#match? @dynamic_import ".*Assembly\\.Load(From|File|"))
            (invocation_expression
                expression: (member_access_expression) @type_load (#match? @type_load ".*Type\\.GetType"))
        `
    case LangElixir:
        queryStr = `
            (call
                target: (identifier) @import (#match? @import "^(import|require|use|alias|Code\\.require_file|Code\\.load_file)$")
                arguments: (arguments (_) @module))
            (call
                target: (dot
                    left: (identifier) @module (#eq? @module "Application")
                    right: (identifier) @method (#eq? @method "ensure_all_started"))
                arguments: (arguments (_) @app))
        `
    case LangElm:
        queryStr = `
            (import_clause
                module_name: (upper_case_qid) @import
                exposing_list: (exposing_list (_) @exposed))
            (module_declaration
                name: (upper_case_qid) @module)
            (port_annotation
                name: (lower_case_identifier) @port)
        `
    case LangGo:
        queryStr = `
            (import_declaration
                (import_spec_list
                    (import_spec
                        path: (interpreted_string_literal) @import
                        name: (package_identifier)? @alias)))
            (import_declaration
                (import_spec
                    path: (interpreted_string_literal) @import
                    name: (package_identifier)? @alias))
            (package_clause
                (package_identifier) @package)
            (call_expression
                function: (selector_expression
                    operand: (identifier) @plugin (#eq? @plugin "plugin")
                    field: (field_identifier) @method (#eq? @method "Open"))
                arguments: (argument_list (interpreted_string_literal) @plugin_path))
        `
    case LangGroovy:
        queryStr = `
            (import_declaration
                (qualified_name) @import)
            (package_declaration
                (qualified_name) @package)
            (method_invocation
                object: (identifier) @class (#eq? @class "Class")
                name: (identifier) @method (#eq? @method "forName")
                arguments: (argument_list (string_literal) @dynamic_import))
            (method_invocation
                object: (method_invocation
                    object: (identifier) @class_loader (#eq? @class_loader "ClassLoader")
                    name: (identifier) @get_method (#eq? @get_method "getSystemClassLoader"))
                name: (identifier) @load_method (#eq? @load_method "loadClass")
                arguments: (argument_list (string_literal) @dynamic_load))
        `
    case LangHcl:
        queryStr = `
            (block
                (identifier) @resource_type
                (string_lit) @resource_name)
            (block
                (identifier) @module_keyword (#eq? @module_keyword "module")
                (string_lit) @module_name
                (block_body (attribute
                    (identifier) @source_attr (#eq? @source_attr "source")
                    (expression (string_lit) @module_source))))
            (attribute
                (identifier) @depends_on (#eq? @depends_on "depends_on")
                (expression (array_expr (_) @dependency)))
        `
    case LangHtml:
        queryStr = `
            (script_element
                (attribute
                    (attribute_name) @attr (#eq? @attr "src")
                    (quoted_attribute_value (attribute_value) @src)))
            (link_element
                (attribute
                    (attribute_name) @attr (#eq? @attr "href")
                    (quoted_attribute_value (attribute_value) @href)))
            (style_element
                (raw_text) @style_content)
            (element
                (start_tag
                    (attribute
                        (attribute_name) @attr (#match? @attr "^(data-import|data-require|data-src|data-module)$")
                        (quoted_attribute_value (attribute_value) @data_import))))
        `
    case LangPhp:
        queryStr = `
            (namespace_name) @namespace
            (namespace_use_clause
                (namespace_name) @use)
            (inclusion_directive
                path: (string_literal) @include)
            (function_call_expression
                function: (name) @func (#match? @func "^(require|require_once|include|include_once|spl_autoload_register)$")
                arguments: (arguments (string_literal)? @path))
            (function_call_expression
                function: (name) @class_func (#eq? @class_func "class_exists")
                arguments: (arguments (string_literal) @class_name))
        `
    case LangScala:
        queryStr = `
            (import_declaration
                (import_expression (stable_identifier) @import))
            (package_declaration
                (stable_identifier) @package)
            (call_expression
                function: (select_expression 
                    (identifier) @class (#eq? @class "Class")
                    (identifier) @method (#eq? @method "forName"))
                arguments: (argument_list (string_literal) @dynamic_class))
            (annotated
                annotation: (annotation
                    name: (simple_identifier) @annotation (#match? @annotation "^(Import|Component|Bean)$")))
        `
    case LangSwift:
        queryStr = `
            (import_declaration
                path: (identifier) @import)
            (module_import_declaration
                (@import) @import_keyword (identifier) @module)
            (function_call_expression
                function: (simple_identifier) @func (#eq? @func "NSClassFromString")
                (tuple (tuple_element (expression (string_literal) @dynamic_class))))
            (attribute
                name: (identifier) @attr (#match? @attr "^(objc|import|exported|_dynamicReplacement)$"))
        `
    default:
        return "", fmt.Errorf("unsupported language for dependency queries: %v", lang)
    }

    return queryStr, nil
}

// ExtractDependencies is the main function to extract dependencies from a file
func ExtractDependencies(path string) ([]string, error) {
    lang := LanguageFromFilePath(path)
    if lang == LangUnknown {
        return nil, fmt.Errorf("unsupported file type: %s", path)
    }
    
    // Parse the file
    parsedFile, err := ParseFile(path)
    if err != nil {
        return nil, fmt.Errorf("could not parse file: %v", err)
    }
    
    if parsedFile == nil || parsedFile.Ast == nil {
        return nil, fmt.Errorf("empty AST for: %s", path)
    }
    
    rootNode := parsedFile.Ast
    var deps []string
    
    queryStr, err := GetDependencyQuery(parsedFile.Language)
    if err != nil {
        return nil, err
    }
    
    // Create and execute the query
    query, err := sitter.NewQuery([]byte(queryStr), parsedFile.Language.Grammar())
    if err != nil {
        return nil, fmt.Errorf("invalid query for %s: %v", path, err)
    }
    
    qc := sitter.NewQueryCursor()
    qc.Exec(query, rootNode)
    
    // Process matches
    for {
        match, ok := qc.NextMatch()
        if !ok {
            break
        }
        
        match = qc.FilterPredicates(match, parsedFile.Source)
        
        for _, capture := range match.Captures {
            dep := capture.Node.Content(parsedFile.Source)
            
            // Clean up the dependency string (remove quotes, etc.)
            dep = cleanDependencyString(dep)
            
            // Skip empty dependencies
            if dep == "" {
                continue
            }
            
            deps = append(deps, dep)
        }
    }
    
    return deps, nil
}

// cleanDependencyString cleans up the extracted dependency string
func cleanDependencyString(dep string) string {
    // Remove quotes for string literals
    dep = strings.Trim(dep, "\"'`")
    
    // Remove any whitespace
    dep = strings.TrimSpace(dep)
    
    return dep
}