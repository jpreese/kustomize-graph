package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	err = GenerateKustomizeGraph(NewKustomizationFile(), currentWorkingDirectory, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(KustomizeGraph.String())
}
