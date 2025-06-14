package analysis

func TaintRun(args ...interface{}) func(*Pass) (any, error) {
	return func(pass *Pass) (any, error) {
		sources := args[0].([]string)
		sinks := args[1].([]string)

		return NewTaintAnalyzer(sources, sinks), nil
	}
}

func NewTaintAnalyzer(sources, sinks []string) *Analyzer {
	analyzer := &Analyzer{}

	return analyzer
}
