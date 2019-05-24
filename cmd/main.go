package main

import (
	"fmt"
	
	"github.com/jpreese/kustomize-graph/pkg/graph"
)

func main() {
	dependencyGraph, err := graph.GenerateKustomizeGraph()
	if err != nil {
		panic(err)
	}

	fmt.Print(dependencyGraph)
}
