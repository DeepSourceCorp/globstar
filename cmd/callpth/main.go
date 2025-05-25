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
	graph, err := analyzer.AnalyzeCodebase("/path/to/code", analysis.ParseGoFile)
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
		ShowThirdPartyCode: true,
		ShowTestFiles: true,
		FocusFunctions: []string{},
	}

	finder := analysis.NewDetailedCallPathFinder(graph, pathConfig)
	paths := finder.FindAllCallPaths()

	for i, path := range paths {
		fmt.Printf("Call Path %d\n", i)
		fmt.Print(path.StringWithContext())
		fmt.Println()
	}
}