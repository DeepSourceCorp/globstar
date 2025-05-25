package analysis

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"os"
)

// Function interface represents a callable unit
type Function interface {
	// identity
	ID() string
	Name() string
	QualifiedName() string

	// Source location
	Filepath() string
	LineNumber() int

	// classification
	IsEntryPoint() bool
	IsThirdParty() bool
}

// Defines a function-caller relationship
type Call interface {
	Caller() Function
	Callee() Function

	CallSiteFile() string
	CallSiteLine() string
}

// interface for the complete call graph structure
type CallGraph interface {
	GetFunctions() []Function
	GetCalls() []Call
	GetEntryPoints() []Function

	GetCallsFrom(function Function) []Call
	GetCallsTo(function Function) []Call

	GetCallers(function Function) []Function
	GetCallees(function Function) []Function
}

type CallGraphConfig struct {
	EntryPointPatterns []string // patterns to match entrypoints
	IncludeThirdParty  bool     // whether to include third party packages
	IncludePaths       []string // paths to include
}

type CallGraphBuilder interface {
	AddFunction(function Function)
	AddCall(call Call)
	SetEntryPoints(functions []Function)
	Build() (CallGraph, error)
}

type BasicFunction struct {
	Id string
	Name_         string
	QualifiedName_ string
	Filepath_     string
	LineNumber_   int
	IsEntryPoint_ bool
	IsThirdParty_ bool
}

func (f *BasicFunction) ID() string {return f.Id}
func (f *BasicFunction) Name() string     { return f.Name_ }
func (f *BasicFunction) QualifiedName() string {return f.QualifiedName_}
func (f *BasicFunction) Filepath() string { return f.Filepath_ }
func (f *BasicFunction) LineNumber() int {return f.LineNumber_}
func (f *BasicFunction) IsEntryPoint() bool {return f.IsEntryPoint_}
func (f *BasicFunction) IsThirdParty() bool {return f.IsThirdParty_}

func (f *BasicFunction) SetEntryPoint(isEntryPoint bool) {
	f.IsEntryPoint_ = isEntryPoint
}

type BasicCall struct {
	CallerFunc Function
	CalleeFunc Function
	SourceFile string
	SourceLine string
}

func (c *BasicCall) Caller() Function {return c.CallerFunc}
func (c *BasicCall) Callee() Function {return c.CalleeFunc}
func (c *BasicCall) CallSiteFile() string {return c.SourceFile}
func (c *BasicCall) CallSiteLine() string {return c.SourceLine}

type BasicCallGraph struct {
	functions map[string]Function
	calls []Call
	callsFrom map[string][]Call
	callsTo map[string][]Call
	entryPoints []Function
}

func NewBasicCallGraph() *BasicCallGraph {
	return &BasicCallGraph{
		functions: make(map[string]Function),
		calls: make([]Call, 0),
		callsFrom: make(map[string][]Call),
		callsTo: make(map[string][]Call),
		entryPoints: make([]Function, 0),
	}
}

func (g *BasicCallGraph) GetFunctions() []Function {
	result := make([]Function, 0, len(g.functions))
	for _, f := range g.functions {
		result = append(result, f)
	}
	return result
}

func (g *BasicCallGraph) GetCalls() []Call {
	return g.calls
}

func (g *BasicCallGraph) GetEntryPoints() []Function {
	return g.entryPoints
}

func (g *BasicCallGraph) GetCallsFrom(function Function) []Call {
	return g.callsFrom[function.ID()]
}

func (g *BasicCallGraph) GetCallsTo(function Function) []Call {
	return g.callsTo[function.ID()]
}

func (g *BasicCallGraph) GetCallers(function Function) []Function {
	calls := g.GetCallsTo(function)
	callerMap := make(map[string]Function)

	for _, call := range calls {
		caller := call.Caller()
		callerMap[caller.ID()] = caller
	}

	callers := make([]Function, 0, len(callerMap))
	for _, caller := range callerMap {
		callers = append(callers, caller)
	}

	return callers
}

func (g *BasicCallGraph) GetCallees(function Function) []Function {
	calls := g.GetCallsFrom(function)
	calleeMap := make(map[string]Function)

	for _, call := range calls {
		callee := call.Callee()
		calleeMap[callee.ID()] = callee
	}

	callees := make([]Function, 0, len(calleeMap))
	for _, callee := range calleeMap {
		callees = append(callees, callee)
	}

	return callees
}

type BasicCallGraphBuilder struct {
	functions map[string]Function
	calls []Call
	entryPoints map[string]bool
	config CallGraphConfig
}

func NewBasicCallGraphBuilder(config CallGraphConfig) *BasicCallGraphBuilder {
	return &BasicCallGraphBuilder{
		functions: make(map[string]Function),
		calls: make([]Call, 0),
		entryPoints: make(map[string]bool),
		config: config,
	}
}

func (b *BasicCallGraphBuilder) AddFunction(function Function) {
	b.functions[function.ID()] = function

	for _, pattern := range b.config.EntryPointPatterns {
		matched, err := filepath.Match(pattern, function.QualifiedName())
		if err == nil && matched {
			b.entryPoints[function.ID()] = true
			break
		}
	}
}

func (b *BasicCallGraphBuilder) AddCall(call Call) {
	b.calls = append(b.calls, call)
}

func (b *BasicCallGraphBuilder) SetEntryPoints(functions []Function) {
	for _, function := range functions {
		b.entryPoints[function.ID()] = true
	}
}

func (b *BasicCallGraphBuilder) Build() (CallGraph, error) {
	graph := NewBasicCallGraph()

	// add functions to graph
	for id, function := range b.functions {
		if function.IsThirdParty() && !b.config.IncludeThirdParty {
			continue
		}

		if basicFunc, ok := function.(*BasicFunction); ok && b.entryPoints[id] {
			basicFunc.SetEntryPoint(true)
		}

		graph.functions[id] = function

		if b.entryPoints[id] {
			graph.entryPoints = append(graph.entryPoints, function)
		}
	}



	// add calls to graph
	for _, call := range b.calls {
		callerId := call.Caller().ID()
		calleeId := call.Callee().ID()



		if _, ok := graph.functions[callerId]; !ok {
			continue
		}
		if _, ok := graph.functions[calleeId]; !ok {

			continue
		}

		graph.calls = append(graph.calls, call)
		graph.callsFrom[callerId] = append(graph.callsFrom[callerId], call)
		graph.callsTo[calleeId] = append(graph.callsTo[calleeId], call)
	}


	return graph, nil

}

type CallGraphGenerator struct {
	config CallGraphConfig
}

func NewCallGraphGenerator(config CallGraphConfig) *CallGraphGenerator {
	return &CallGraphGenerator{
		config: config,
	}
}

func (g *CallGraphGenerator) CreateCallGraph(functions []Function, calls []Call) (CallGraph, error) {
	builder := NewBasicCallGraphBuilder(g.config)

	for _, function := range functions {
		builder.AddFunction(function)
	}

	for _, call := range calls {
		builder.AddCall(call)
	}

	return builder.Build()
}

// handle parsing a codebase to generate a call graph
type CodebaseAnalyzer struct {
	config CallGraphConfig
}

func NewCodeBaseAnalyzer(config CallGraphConfig) *CodebaseAnalyzer {
	return &CodebaseAnalyzer{
		config: config,
	}
}

func (a *CodebaseAnalyzer) AnalyzeCodebase(rootpath string, parseFile func(string) ([]Function, []Call, error)) (CallGraph, error) {
	allFunctions := make([]Function, 0)
	allCalls := make([]Call, 0)

	err := filepath.WalkDir(rootpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !a.shouldIncludeFile(path) {
			return nil
		}

		functions, calls, parseErr := parseFile(path)
		if parseErr != nil {
			fmt.Printf("warning: failed to parse %s: %v\n", path, parseErr)
			return nil
		}

		allFunctions = append(allFunctions, functions...)
		allCalls = append(allCalls, calls...)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error analyzing codebase at %s: %w", rootpath, err)
	}

	return a.buildCallGraph(allFunctions, allCalls)
}

func (a *CodebaseAnalyzer) ExtractFunctionsAndCallsFromPaths(paths []string, parseFile func(string) ([]Function, []Call, error)) ([]Function, []Call, error) {
	allFunctions := make([]Function, 0)
	allCalls := make([]Call, 0)

	for _, rootpath := range paths {
		fmt.Printf("Analyzing path: %s\n", rootpath)
		
		// Check if path exists
		if _, err := os.Stat(rootpath); os.IsNotExist(err) {
			fmt.Printf("Warning: path %s doesn't exist, skipping...\n", rootpath)
			continue
		}

		pathFunctions, pathCalls, err := a.extractFromSinglePath(rootpath, parseFile)
		if err != nil {
			fmt.Printf("Warning: failed to analyze %s: %v\n", rootpath, err)
			continue
		}

		fmt.Printf("  Found %d functions, %d calls\n", len(pathFunctions), len(pathCalls))
		
		allFunctions = append(allFunctions, pathFunctions...)
		allCalls = append(allCalls, pathCalls...)
	}

	return allFunctions, allCalls, nil
}

func (a *CodebaseAnalyzer) extractFromSinglePath(rootpath string, parseFile func(string) ([]Function, []Call, error)) ([]Function, []Call, error) {
	allFunctions := make([]Function, 0)
	allCalls := make([]Call, 0)

	err := filepath.WalkDir(rootpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !a.shouldIncludeFile(path) {
			return nil
		}

		functions, calls, parseErr := parseFile(path)
		if parseErr != nil {
			fmt.Printf("warning: failed to parse %s: %v\n", path, parseErr)
			return nil
		}

		allFunctions = append(allFunctions, functions...)
		allCalls = append(allCalls, calls...)

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error analyzing codebase at %s: %w", rootpath, err)
	}
	return allFunctions, allCalls, nil
}


func (a *CodebaseAnalyzer) shouldIncludeFile(filepath string) bool {
	if len(a.config.IncludePaths) > 0 {
		included := false
		for _, includePath := range a.config.IncludePaths {
			if strings.HasPrefix(filepath, includePath) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	excludePaths := []string{
		"/.git/",
		"/vendor/",
		"__pycache__",
		"/target/",
		"/.idea/",
		"/.vscode/",
	}

	for _, excludePath := range excludePaths {
		if strings.Contains(filepath, excludePath) {
			return false
		}
	}

	return true
}

func (a *CodebaseAnalyzer) buildCallGraph(functions []Function, calls []Call) (CallGraph, error) {
	generator := NewCallGraphGenerator(a.config)
	return generator.CreateCallGraph(functions, calls)
}