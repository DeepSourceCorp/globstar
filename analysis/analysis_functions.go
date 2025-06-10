package analysis

func TaintRun(args ...interface{}) Analyzer {
	sources := args[0].([]string)
	sinks := args[1].([]string)

	analyzer := NewTaintAnalyzer(sources, sinks)
	return analyzer
}

func NewTaintAnalyzer(sources, sinks []string) Analyzer {
	return Analyzer{
		Name: "taint_analyzer",
	}
}


