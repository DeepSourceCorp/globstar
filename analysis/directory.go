package analysis

var RegisteredAnalysisFunctions = []AnalysisFunction{}

func InitializeAnalysisFunction(fn AnalysisFunction) AnalysisFunction {
	switch fn.Name {
	case "taint":
		fn.Run = TaintRun
	default:
		break
	}
	RegisteredAnalysisFunctions = append(RegisteredAnalysisFunctions, fn)
	return fn
}
