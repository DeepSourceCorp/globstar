package analysis

var RegisteredAnalysisFunctions = []AnalysisFunction{}

func InitializeAnalysisFunction(fn AnalysisFunction) AnalysisFunction {
	switch fn.Name {
	case "taint":
		fn.Run = TaintRun
	}

	RegisteredAnalysisFunctions = append(RegisteredAnalysisFunctions, fn)
	return fn
}

func TaintRun(args ...interface{}) (Analyzer, error) {
	sources := args[0].([]string)
	sinks := args[1].([]string)

	analyzer := NewTaintAnalyzer(sources, sinks)
	return analyzer, nil
}

func NewTaintAnalyzer(sources, sinks []string) Analyzer {
	return Analyzer{
		Name: "taint_analyzer",
	}
}
