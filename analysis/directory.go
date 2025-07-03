package analysis

import "fmt"

type FunctionDirectory struct {
	// Different languages can have the same type of Analysis function.
	// This first maps the Type of function-mode, then the specific language implementation
	Pool map[string]map[Language]*AnalysisFunction
}

var AnalysisFuncDirectory = &FunctionDirectory{
	Pool: make(map[string]map[Language]*AnalysisFunction),
}

func (fd *FunctionDirectory) AddToDirectory(ana *Analyzer) error {
	anaFunc := fd.Pool[ana.Name]
	if anaFunc == nil {
		return fmt.Errorf("%s method is not supported", ana.Name)
	}
	anaFunc[ana.Language].Analyzer = ana
	return nil
}
