package analysis

import (
	"bufio"
	"os"
	"strings"
	"fmt"
)

// call path structure
type CallPath struct {
	Functions []Function
	Calls     []Call
	Depth     int
}

func (cp *CallPath) String() string {
	if len(cp.Functions) == 0 {
		return "empty path"
	}

	var parts []string
	for _, function := range cp.Functions {
		parts = append(parts, function.QualifiedName())
	}

	return strings.Join(parts, " -> ")
}

type CallPathFinder struct {
	graph CallGraph
}

func NewCallPathFinder(graph CallGraph) *CallPathFinder {
	return &CallPathFinder{graph: graph}
}

func (cpf *CallPathFinder) FindPathsFromEntryPoints(targetFunc Function, maxDepth int) []CallPath {
	var allPaths []CallPath

	entryPoints := cpf.graph.GetEntryPoints()
	for _, entryPoint := range entryPoints {
		paths := cpf.FindPaths(entryPoint, targetFunc, maxDepth)
		allPaths = append(allPaths, paths...)
	}

	return allPaths
}

func (cpf *CallPathFinder) FindPaths(sourceFunc, targetFunc Function, maxDepth int) []CallPath {
	var paths []CallPath
	visited := make(map[string]bool)
	currentPath := CallPath{
		Functions: []Function{sourceFunc},
		Calls:     []Call{},
		Depth:     0,
	}

	cpf.findPathsRecursive(sourceFunc, targetFunc, currentPath, visited, maxDepth, &paths)
	return paths
}

func (cpf *CallPathFinder) findPathsRecursive(current, target Function, currentPath CallPath, visited map[string]bool, maxDepth int, paths *[]CallPath) {
	if current.ID() == target.ID() {
		pathCopy := CallPath{
			Functions: make([]Function, len(currentPath.Functions)),
			Calls:     make([]Call, len(currentPath.Calls)),
			Depth:     currentPath.Depth,
		}

		copy(pathCopy.Functions, currentPath.Functions)
		copy(pathCopy.Calls, currentPath.Calls)
		*paths = append(*paths, pathCopy)
		return
	}

	if currentPath.Depth >= maxDepth {
		return
	}

	visited[current.ID()] = true
	defer func() { delete(visited, current.ID()) }()

	callsFromCurrent := cpf.graph.GetCallsFrom(current)

	for _, call := range callsFromCurrent {
		callee := call.Callee()

		if visited[callee.ID()] {
			continue
		}

		newPath := CallPath{
			Functions: make([]Function, len(currentPath.Functions)),
			Calls:     make([]Call, len(currentPath.Calls)),
			Depth:     currentPath.Depth + 1,
		}

		copy(newPath.Functions, currentPath.Functions)
		copy(newPath.Calls, currentPath.Calls)

		cpf.findPathsRecursive(callee, target, newPath, visited, maxDepth, paths)
	}
}

func (cpf *CallPathFinder) FindCalleePaths(sourceFunc Function, maxDepth int) []CallPath {
	var allPaths []CallPath
	visited := make(map[string]bool)
	currentPath := CallPath{
		Functions: []Function{sourceFunc},
		Calls:     []Call{},
		Depth:     0,
	}

	cpf.findAllPathsFromFunction(sourceFunc, currentPath, visited, maxDepth, &allPaths)
	return allPaths
}

func (cpf *CallPathFinder) findAllPathsFromFunction(current Function, currentPath CallPath, visited map[string]bool, maxDepth int, paths *[]CallPath) {
	if len(currentPath.Functions) > 1 {
		pathCopy := CallPath{
			Functions: make([]Function, len(currentPath.Functions)),
			Calls:     make([]Call, len(currentPath.Calls)),
			Depth:     currentPath.Depth,
		}

		copy(pathCopy.Functions, currentPath.Functions)
		copy(pathCopy.Calls, currentPath.Calls)
		*paths = append(*paths, pathCopy)
	}

	if currentPath.Depth >= maxDepth {
		return
	}

	visited[current.ID()] = true
	defer func() { delete(visited, current.ID()) }()

	callsFromCurrent := cpf.graph.GetCallsFrom(current)

	for _, call := range callsFromCurrent {
		callee := call.Callee()

		if visited[callee.ID()] {
			continue
		}

		newPath := CallPath{
			Functions: make([]Function, len(currentPath.Functions)),
			Calls:     make([]Call, len(currentPath.Calls)),
			Depth:     currentPath.Depth + 1,
		}

		copy(newPath.Functions, currentPath.Functions)
		copy(newPath.Calls, currentPath.Calls)

		newPath.Functions = append(newPath.Functions, callee)
		newPath.Calls = append(newPath.Calls, call)

		cpf.findAllPathsFromFunction(callee, newPath, visited, maxDepth, paths)
	}
}

type DetailedCallPath struct {
	CallPath
	CodeContext map[string]CodeSnippet
	PathType    string
}

type CodeSnippet struct {
	FilePath     string
	FunctionName string
	TargetLine   int
	StartLine    int
	EndLine      int
	Lines        []CodeLine
}

type CodeLine struct {
	Number   int
	Content  string
	IsTarget bool
}

type CallPathConfig struct {
	ContextLinesBefore int
	ContextLinesAfter  int
	MaxPathDepth       int
	ShowTestFiles      bool
	ShowThirdPartyCode bool
	FocusFunctions     []string
	ExcludePaths       []string
}


type DetailedCallPathFinder struct {
	graph CallGraph
	config CallPathConfig
}

func NewDetailedCallPathFinder(graph CallGraph, config CallPathConfig) *DetailedCallPathFinder {
	if config.ContextLinesBefore == 0 {
		config.ContextLinesBefore = 3
	}

	if config.ContextLinesAfter == 0 {
		config.ContextLinesAfter = 3
	}

	if config.MaxPathDepth == 0 {
		config.MaxPathDepth = 10
	}

	return &DetailedCallPathFinder{
		graph: graph,
		config: config,
	}
}

func (dcpf *DetailedCallPathFinder) FindAllCallPaths() []DetailedCallPath {
	var allPaths []DetailedCallPath

	entryPoints := dcpf.graph.GetEntryPoints()
	if len(entryPoints) == 0 {
		// if no entrypoints defined, use functions that are not called by others
		entryPoints = dcpf.findRootFunctions()
	}

	finder := NewCallPathFinder(dcpf.graph)

	// for each entrypoint, find all paths it can reach
	for _, entryPoint := range entryPoints {
		paths := finder.FindCalleePaths(entryPoint, dcpf.config.MaxPathDepth)

		for _, path := range paths {
			detailedPath := dcpf.createDetailedCallPath(path)
			if detailedPath != nil && dcpf.shouldIncludePath(*detailedPath) {
				allPaths = append(allPaths, *detailedPath)
			}
		}
	}

	return allPaths
}

func (dcpf *DetailedCallPathFinder) findRootFunctions() []Function {
	var rootFunctions []Function

	for _, function := range dcpf.graph.GetFunctions() {
		callers := dcpf.graph.GetCallers(function)
		if len(callers) == 0 {
			rootFunctions = append(rootFunctions, function)
		}
	}

	return rootFunctions
}

func (dcpf *DetailedCallPathFinder) createDetailedCallPath(callPath CallPath) *DetailedCallPath {
	if len(callPath.Functions) == 0 {
		return nil
	}

	pathType := determinePathType(callPath.Functions)

	codeContext := dcpf.extractCodeContext(callPath.Functions, callPath.Calls)

	return &DetailedCallPath{
		CallPath: callPath,
		CodeContext: codeContext,
		PathType: pathType,
	}
}

func determinePathType(functions []Function) string {
	hasInternal := false
	hasThirdParty := false

	for _, function := range functions {
		if function.IsThirdParty() {
			hasThirdParty = true
		} else {
			hasInternal = true
		}
	}

	if hasInternal && hasThirdParty {
		return "mixed"
	} else if hasThirdParty {
		return "third_party"
	} else {
		return "internal"
	}
}

func (dcpf *DetailedCallPathFinder) extractCodeContext(functions []Function, calls []Call) map[string]CodeSnippet {
	context := make(map[string]CodeSnippet)

	for _, function := range functions {
		if function.IsThirdParty() && !dcpf.config.ShowThirdPartyCode {
			continue
		}

		if !dcpf.config.ShowTestFiles {
			continue
		}

		snippet := dcpf.getCodeSnippet(function, function.LineNumber())
		if snippet != nil {
			context[function.ID()] = *snippet
		}
	}

	return context
}

func (dcpf *DetailedCallPathFinder) getCodeSnippet(function Function, targetLine int) *CodeSnippet {
	filePath := function.Filepath()

	if targetLine <= 0 || filePath == "" || filePath == "unknown" {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var allLines []string

	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if len(allLines) == 0 || targetLine > len(allLines) {
		return nil
	}

	startLine := max(1, targetLine - dcpf.config.ContextLinesBefore)
	endLine := min(len(allLines), targetLine + dcpf.config.ContextLinesAfter)

	// extract lines in range
	var lines []CodeLine
	for i := startLine; i <= endLine; i++ {
		if i - 1 < len(allLines) {
			lines = append(lines, CodeLine{
				Number: i,
				Content: allLines[i - 1],
				IsTarget: i == targetLine,
			})
		}
	}

	return &CodeSnippet{
		FilePath: filePath,
		FunctionName: function.Name(),
		TargetLine: targetLine,
		StartLine: startLine,
		EndLine: endLine,
		Lines: lines,
	}
}


func (dcpf *DetailedCallPathFinder) shouldIncludePath(path DetailedCallPath) bool {
	if len(dcpf.config.FocusFunctions) > 0 {
		found := false
		for _, function := range path.Functions {
			for _, focus := range dcpf.config.FocusFunctions {
				if strings.Contains(strings.ToLower(function.QualifiedName()), strings.ToLower(focus)) {
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return false
		}
	}

	for _, excludePattern := range dcpf.config.ExcludePaths {
		for _, function := range path.Functions {
			if strings.Contains(strings.ToLower(function.Filepath()), strings.ToLower(excludePattern)) {
				return false
			}
		}
	}

	return true
}

func (dcp *DetailedCallPath) String() string {
	var result strings.Builder
	
	result.WriteString("=== Call Path ===\n")
	result.WriteString(fmt.Sprintf("Type: %s\n", dcp.PathType))
	result.WriteString(fmt.Sprintf("Depth: %d\n", dcp.Depth))
	result.WriteString("\n")
	
	// Show the call sequence
	for i, function := range dcp.Functions {
		nodeNum := i + 1
		codeType := "Internal code"
		if function.IsThirdParty() {
			codeType = "Third-party code"
		}
		
		result.WriteString(fmt.Sprintf("%d. Function: %s, File: %s, Line: %d (%s)\n",
			nodeNum,
			function.Name(),
			function.Filepath(),
			function.LineNumber(),
			codeType,
		))
	}
	
	result.WriteString("\n")
	return result.String()
}