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
