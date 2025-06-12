package analysis

type TaintAnalyzer interface {
	GetAnalyzer(sources, sinks []string) Analyzer
}

func TaintRun(args ...interface{}) Analyzer {
	sources := args[0].([]string)
	sinks := args[1].([]string)

	analyzer := NewTaintAnalyzer(sources, sinks)
	return analyzer
}

func NewTaintAnalyzer(sources, sinks []string) Analyzer {
	var taintAnalyzer TaintAnalyzer
	analyzer := taintAnalyzer.GetAnalyzer(sources, sinks)
	return analyzer
}
