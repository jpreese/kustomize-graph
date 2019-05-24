package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jpreese/kustomize-graph/pkg/graph"
	"github.com/jpreese/kustomize-graph/pkg/kustomizationfile"
)

func main() {

	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	dependencyGraph, err := graph.GenerateKustomizeGraph(kustomizationfile.DefaultContext(), workingDirectory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(dependencyGraph.String())
}
