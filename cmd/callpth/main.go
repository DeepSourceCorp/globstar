package main

import (
	"fmt"
	"log"

	"globstar.dev/analysis"
)

func main() {
	config := analysis.CallGraphConfig{
		EntryPointPatterns: []string{".main", "main.main", "*.main"},
		IncludeThirdParty: true,
	}

	// paths for codebase and third party package
	paths := []string {}

	analyzer := analysis.NewCodeBaseAnalyzer(config)
	allFunctions, allCalls, err := analyzer.ExtractFunctionsAndCallsFromPaths(paths, analysis.ParseGoFile)
	if err != nil {
		panic(err)
	}
	
	resolved := analysis.ResolveCallsToFunctions(allCalls, allFunctions)

	generator := analysis.NewCallGraphGenerator(config)
	graph, err := generator.CreateCallGraph(allFunctions, resolved)
	if err != nil {
		log.Fatalf("error creating combined graph: %v", err)
	}

	functions := graph.GetFunctions()
	calls := graph.GetCalls()
	entryPoints := graph.GetEntryPoints()

	
	fmt.Printf("Final graph: %d functions, %d calls, %d entry points\n", 
		len(functions), len(calls), len(entryPoints))


	pathConfig := analysis.CallPathConfig{
		ContextLinesBefore: 3,
		ContextLinesAfter: 3,
		MaxPathDepth: 4,
		ShowThirdPartyCode: true,
		ShowTestFiles: true,
		FocusFunctions: []string{"createServer"},
	}

	finder := analysis.NewDetailedCallPathFinder(graph, pathConfig)
	callpaths := finder.FindAllCallPaths()

	for i, callpath := range callpaths {
		fmt.Printf("Call Path %d\n", i)
		fmt.Print(callpath.StringWithContext())
		fmt.Println()
	}
}