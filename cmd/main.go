package main

import (
	"fmt"

	"github.com/jpreese/kustomize-graph/pkg/kustomizationgraph"
)

func main() {
	graph, err := kustomizationgraph.New("main").Generate()
	if err != nil {
		panic(err)
	}

	fmt.Print(graph)
}
