package analysis

import "fmt"

type FunctionDirectory struct {
	Pool map[string]*AnalysisFunction
}

var AnalysisFuncDirectory = &FunctionDirectory{
	Pool: make(map[string]*AnalysisFunction),
}

func (fd *FunctionDirectory) AddToDirectory(ana *Analyzer) error {
	anaFunc := fd.Pool[ana.Name]
	if anaFunc == nil {
		return fmt.Errorf("%s method is not supported", ana.Name)
	}
	anaFunc.Analyzer = ana
	return nil
}
