package analysis

type DependencyGraph struct {
	dependencies map[string][]string // file -> dependencies
	dependents map[string][]string // file -> files that depend on it
}

func NewDependencyGraph() *DependencyGraph {
	dg := &DependencyGraph{
		dependencies: map[string][]string{},
		dependents: map[string][]string{},
	}

	return dg
}


func (dg *DependencyGraph) AddDependency(file, dependency string){
	dg.dependencies[file] = append(dg.dependencies[file], dependency)
	dg.dependents[dependency] = append(dg.dependents[dependency], file)
}


func (g *DependencyGraph) GetAffectedFiles(changedFiles []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	var visit func(string)
	visit = func(file string) {
		// if the file has already been analyzed, skip it 
		if seen[file]{
			return
		}

		seen[file] = true
		result = append(result, file)

		for _, dependent := range g.dependents[file] {
			visit(dependent)
		}

	}

	for _, file := range changedFiles {
		visit(file)
	}

	return result
}