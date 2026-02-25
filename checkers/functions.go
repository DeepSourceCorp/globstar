package checkers

import (
	"globstar.dev/analysis"
	"globstar.dev/checkers/javascript"
)

func InitializeAnalysisFunctionDirectory(name string, language analysis.Language) {
	// Find a way to automate the registration of analyzers than adding them manually
	switch name {
	case "taint":
		switch language {
		case analysis.LangJs:
			javascript.JsTaintAnalyzer()
		}
	default:
		return
	}
}
