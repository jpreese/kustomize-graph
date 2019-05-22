package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jpreese/kustomize-graph/pkg/graph"
	"github.com/jpreese/kustomize-graph/pkg/kustomize"
)

func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	kustomizationFile, err := kustomize.NewKustomizationFile().Get(workingDirectory)
	if err != nil {
		log.Fatal("Unable to get kustomization.yaml file from directory %s", workingDirectory)
		return
	}

	dependencyGraph, err := graph.GenerateKustomizeGraph(graph.New(), kustomizationFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(dependencyGraph)
}
