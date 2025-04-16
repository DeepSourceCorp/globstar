package analysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"

	treeSitterBash "github.com/smacker/go-tree-sitter/bash"
	treeSitterCsharp "github.com/smacker/go-tree-sitter/csharp"
	treeSitterCss "github.com/smacker/go-tree-sitter/css"
	treeSitterDockerfile "github.com/smacker/go-tree-sitter/dockerfile"
	treeSitterElixir "github.com/smacker/go-tree-sitter/elixir"
	treeSitterElm "github.com/smacker/go-tree-sitter/elm"
	treeSitterGo "github.com/smacker/go-tree-sitter/golang"
	treeSitterGroovy "github.com/smacker/go-tree-sitter/groovy"
	treeSitterHcl "github.com/smacker/go-tree-sitter/hcl"
	treeSitterHtml "github.com/smacker/go-tree-sitter/html"
	treeSitterJava "github.com/smacker/go-tree-sitter/java"
	treeSitterKotlin "github.com/smacker/go-tree-sitter/kotlin"
	treeSitterLua "github.com/smacker/go-tree-sitter/lua"
	treeSitterOCaml "github.com/smacker/go-tree-sitter/ocaml"
	treeSitterPhp "github.com/smacker/go-tree-sitter/php"
	treeSitterPy "github.com/smacker/go-tree-sitter/python"
	treeSitterRuby "github.com/smacker/go-tree-sitter/ruby"
	treeSitterRust "github.com/smacker/go-tree-sitter/rust"
	treeSitterScala "github.com/smacker/go-tree-sitter/scala"
	treeSitterSql "github.com/smacker/go-tree-sitter/sql"
	treeSitterSwift "github.com/smacker/go-tree-sitter/swift"
	treeSitterTsx "github.com/smacker/go-tree-sitter/typescript/tsx"
	treeSitterTs "github.com/smacker/go-tree-sitter/typescript/typescript"
)

// ParseResult is the result of parsing a file.
type ParseResult struct {
	// Ast is the root node of the tree-sitter parse-tree
	// representing this file
	Ast *sitter.Node
	// Source is the raw source code of the file
	Source []byte
	// FilePath is the path to the file that was parsed
	FilePath string
	// Language is the tree-sitter language used to parse the file
	TsLanguage *sitter.Language
	// Language is the language of the file
	Language Language
	// ScopeTree represents the scope hierarchy of the file.
	// Can be nil if scope support for this language has not been implemented yet.
	ScopeTree *ScopeTree
}

type Language int

const (
	LangUnknown Language = iota
	LangPy
	LangJs  // vanilla JS and JSX
	LangTs  // TypeScript (not TSX)
	LangTsx // TypeScript with JSX extension
	LangJava
	LangRuby
	LangRust
	LangYaml
	LangCss
	LangDockerfile
	LangMarkdown
	LangSql
	LangKotlin
	LangOCaml
	LangLua
	LangBash
	LangCsharp
	LangElixir
	LangElm
	LangGo
	LangGroovy
	LangHcl
	LangHtml
	LangPhp
	LangScala
	LangSwift
)

// tsGrammarForLang returns the tree-sitter grammar for the given language.
// May return `nil` when `lang` is `LangUnkown`.
func (lang Language) Grammar() *sitter.Language {
	switch lang {
	case LangPy:
		return treeSitterPy.GetLanguage()
	case LangJs:
		return treeSitterTsx.GetLanguage() // Use TypeScript's JSX grammar for JS/JSX
	case LangTs:
		return treeSitterTs.GetLanguage()
	case LangTsx:
		return treeSitterTsx.GetLanguage()
	case LangJava:
		return treeSitterJava.GetLanguage()
	case LangRuby:
		return treeSitterRuby.GetLanguage()
	case LangRust:
		return treeSitterRust.GetLanguage()
	case LangSql:
		return treeSitterSql.GetLanguage()
	case LangKotlin:
		return treeSitterKotlin.GetLanguage()
	case LangCss:
		return treeSitterCss.GetLanguage()
	case LangOCaml:
		return treeSitterOCaml.GetLanguage()
	case LangLua:
		return treeSitterLua.GetLanguage()
	case LangDockerfile:
		return treeSitterDockerfile.GetLanguage()
	case LangBash:
		return treeSitterBash.GetLanguage()
	case LangCsharp:
		return treeSitterCsharp.GetLanguage()
	case LangElixir:
		return treeSitterElixir.GetLanguage()
	case LangElm:
		return treeSitterElm.GetLanguage()
	case LangGo:
		return treeSitterGo.GetLanguage()
	case LangGroovy:
		return treeSitterGroovy.GetLanguage()
	case LangHcl:
		return treeSitterHcl.GetLanguage()
	case LangHtml:
		return treeSitterHtml.GetLanguage()
	case LangPhp:
		return treeSitterPhp.GetLanguage()
	case LangScala:
		return treeSitterScala.GetLanguage()
	case LangSwift:
		return treeSitterSwift.GetLanguage()
	default:
		return nil
	}
}

// NOTE(@injuly): TypeScript and TSX have to parsed with DIFFERENT
// grammars. Otherwise, because an expression like `<Foo>bar` is
// parsed as a (legacy) type-cast in TS, but a JSXElement in TSX.
// See: https://facebook.github.io/jsx/#prod-JSXElement

// LanguageFromFilePath returns the Language of the file at the given path
// returns `LangUnkown` if the language is not recognized (e.g: `.txt` files).
func LanguageFromFilePath(path string) Language {
	ext := filepath.Ext(path)
	switch ext {
	case ".py":
		return LangPy
		// TODO: .jsx and .js can both have JSX syntax -_-
	case ".js", ".jsx":
		return LangJs
	case ".ts":
		return LangTs
	case ".tsx":
		return LangTs
	case ".java":
		return LangJava
	case ".rb":
		return LangRuby
	case ".rs":
		return LangRust
	case ".css":
		return LangCss
	case ".Dockerfile":
		return LangDockerfile
	case ".sql":
		return LangSql
	case ".kt":
		return LangKotlin
	case ".ml":
		return LangOCaml
	case ".lua":
		return LangLua
	case ".sh":
		return LangBash
	case ".cs":
		return LangCsharp
	case ".ex":
		return LangElixir
	case ".elm":
		return LangElm
	case ".go":
		return LangGo
	case ".groovy":
		return LangGroovy
	case ".tf":
		return LangHcl
	case ".html":
		return LangHtml
	case ".php":
		return LangPhp
	case ".scala":
		return LangScala
	case ".swift":
		return LangSwift
	default:
		return LangUnknown
	}
}

func GetExtFromLanguage(lang Language) string {
	switch lang {
	case LangPy:
		return ".py"
	case LangJs:
		return ".js"
	case LangTs:
		return ".ts"
	case LangTsx:
		return ".tsx"
	case LangJava:
		return ".java"
	case LangRuby:
		return ".rb"
	case LangRust:
		return ".rs"
	case LangYaml:
		return ".yaml"
	case LangCss:
		return ".css"
	case LangDockerfile:
		return ".Dockerfile"
	case LangSql:
		return ".sql"
	case LangKotlin:
		return ".kt"
	case LangOCaml:
		return ".ml"
	case LangLua:
		return ".lua"
	case LangBash:
		return ".sh"
	case LangCsharp:
		return ".cs"
	case LangElixir:
		return ".ex"
	case LangElm:
		return ".elm"
	case LangGo:
		return ".go"
	case LangGroovy:
		return ".groovy"
	case LangHcl:
		return ".tf"
	case LangHtml:
		return ".html"
	case LangPhp:
		return ".php"
	case LangScala:
		return ".scala"
	case LangSwift:
		return ".swift"
	default:
		return ""
	}
}

func Parse(filePath string, source []byte, language Language, grammar *sitter.Language) (*ParseResult, error) {
	ast, err := sitter.ParseCtx(context.Background(), source, grammar)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	scopeTree := MakeScopeTree(language, ast, source)
	parseResult := &ParseResult{
		Ast:        ast,
		Source:     source,
		FilePath:   filePath,
		TsLanguage: grammar,
		Language:   language,
		ScopeTree:  scopeTree,
	}

	return parseResult, nil
}

// ParseFile parses the file at the given path using the appropriate
// tree-sitter grammar.
func ParseFile(filePath string) (*ParseResult, error) {
	lang := LanguageFromFilePath(filePath)
	grammar := lang.Grammar()
	if grammar == nil {
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Parse(filePath, source, lang, grammar)
}

func GetEscapedCommentIdentifierFromPath(path string) string {
	lang := LanguageFromFilePath(path)
	switch lang {
	case LangJs, LangTs, LangTsx, LangJava, LangRust, LangCss, LangMarkdown, LangKotlin, LangCsharp, LangGo, LangGroovy, LangPhp, LangScala, LangSwift:
		return "\\/\\/"
	case LangPy, LangLua, LangBash, LangRuby, LangYaml, LangDockerfile, LangElixir, LangHcl:
		return "#"
	case LangSql, LangElm:
		return "--"
	case LangHtml:
		return "<\\!--"
	case LangOCaml:
		return "\\(\\*"
	default:
		return ""
	}
}
