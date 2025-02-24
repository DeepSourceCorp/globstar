package analysis

import (
	"fmt"

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
			(import_from_statement)
				module_name: (dotted_name) @from)
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
		`
	case LangJava:
		queryStr = `
			(import_declaration
				(scoped_identifier) @import)
			(package_declaration
				(scoped_identifier) @package)
		`
	case LangRuby:
		queryStr = `
			(call
				method: (identifier) @method (#match? @method "^(require|require_relative|load|autoload)$")
				arguments: (argument_list (string) @path))
			(class (constant) @class)
			(module (constant) @module)
		`
	case LangRust:
		queryStr = `
			(use_declaration
				path: (path) @import)
			(extern_crate_declaration
				name: (identifier) @crate)
		`
	case LangSql:
		// SQL doesn't typically have imports, but we can look for includes
		queryStr = `
			(include_statement
				file_name: (_) @include)
		`
	case LangKotlin:
		queryStr = `
			(import_header
				identifier: (identifier) @import)
			(package_header
				identifier: (identifier) @package)
		`
	case LangCss:
		queryStr = `
			(import_statement
				url: (string_value) @import)
			(at_rule
				name: (at_keyword) @rule (#eq? @rule "@import")
				value: (prelude (string_value) @path))
		`
	case LangOCaml:
		queryStr = `
			(open_statement
				module_path: (module_path) @open)
			(include_statement
				module_expression: (module_path) @include)
		`
	case LangLua:
		queryStr = `
			(function_call
				name: (identifier) @require (#eq? @require "require")
				arguments: (arguments (string) @path))
		`
	case LangDockerfile:
		queryStr = `
			(from_instruction
				image: (image_spec (image_name) @from))
		`
	case LangBash:
		queryStr = `
			(command
				name: (command_name) @command (#match? @command "^(source|\\.)$")
				argument: (word) @source_path)
		`
	case LangCsharp:
		queryStr = `
			(using_directive
				name: (qualified_name) @using)
			(namespace_declaration
				name: (qualified_name) @namespace)
		`
	case LangElixir:
		queryStr = `
			(call
				target: (identifier) @import (#match? @import "^(import|require|use|alias)$")
				arguments: (arguments (_) @module))
		`
	case LangElm:
		queryStr = `
			(import_clause
				module_name: (upper_case_qid) @import)
		`
	case LangGo:
		queryStr = `
			(import_declaration
				(import_spec name: (package_identifier)? path: (interpreted_string_literal) @import))
			(package_clause
				(package_identifier) @package)
		`
	case LangGroovy:
		queryStr = `
			(import_declaration
				(qualified_name) @import)
			(package_declaration
				(qualified_name) @package)
		`
	case LangHcl:
		queryStr = `
			(block
				(identifier) @resource_type
				(string_lit) @resource_name)
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
		`
	case LangPhp:
		queryStr = `
			(namespace_name) @namespace
			(namespace_use_clause
				(namespace_name) @use)
			(inclusion_directive
				path: (string_literal) @include)
		`
	case LangScala:
		queryStr = `
			(import_declaration
				(import_expression (stable_identifier) @import))
			(package_declaration
				(stable_identifier) @package)
		`
	case LangSwift:
		queryStr = `
			(import_declaration
				path: (identifier) @import)
		`
	default:
		return "", fmt.Errorf("unsupported language for dependency queries: %v", lang)
	}

	return queryStr, nil
}


// Function that extracts and stores the file-path of the dependency of the passed in file.
func ExtractDependency(path string) ([]string, error) {
	lang := LanguageFromFilePath(path)
	// parsing the file and getting it's AST
	parsedFile, err := ParseFile(path)

	if err != nil {
		return nil, fmt.Errorf("Could not parse file, err: %v", err)
	}

	rootNode := parsedFile.Ast

	var deps []string
	queryStr, err := GetDependencyQuery(parsedFile.Language)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// using the import patterns to detect dependency in a file
	query, err := sitter.NewQuery([]byte(queryStr), lang.Grammar())

	qc := sitter.NewQueryCursor()
	qc.Exec(query, rootNode)

	for {
		match, ok := qc.NextMatch()

		if !ok {
			break
		}

		for _, capture := range match.Captures {
			dep := capture.Node.Content(parsedFile.Source)
			deps = append(deps, dep)

		}
	}

	return deps, nil
}

