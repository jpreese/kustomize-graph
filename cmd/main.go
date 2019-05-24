package main

import (
	"fmt"
	"os"

	"github.com/jpreese/kustomize-graph/pkg/graph"
)

func main() {
	dependencyGraph, err := graph.GenerateKustomizeGraph()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(dependencyGraph.String())
}
