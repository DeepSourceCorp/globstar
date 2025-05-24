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

	analyzer := analysis.NewCodeBaseAnalyzer(config)
	graph, err := analyzer.AnalyzeCodebase("/home/okabe/graveyard/simple-mock", analysis.ParseGoFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	functions := graph.GetFunctions()
	calls := graph.GetCalls()
	entryPoints := graph.GetEntryPoints()
	
	fmt.Printf("Found %d functions\n", len(functions))
	fmt.Printf("Found %d calls\n", len(calls))
	fmt.Printf("Found %d entry points\n", len(entryPoints))

	pathConfig := analysis.CallPathConfig{
		ContextLinesBefore: 3,
		ContextLinesAfter: 3,
		MaxPathDepth: 15,
		ShowThirdPartyCode: false,
		FocusFunctions: []string{"transformData", "saveResult"},
	}

	finder := analysis.NewDetailedCallPathFinder(graph, pathConfig)
	paths := finder.FindAllCallPaths()

	for i, path := range paths {
		fmt.Printf("Path %d: %s\n", i+1, path.CallPath.String())
	}
}