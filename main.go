package main

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type KustomizationFileStructure struct {
	Bases     []string `yaml:"bases"`
	Resources []string `yaml:"resources"`
}

var graphAst, _ = gographviz.ParseString(`digraph main {}`)
var graph = gographviz.NewGraph()

func main() {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	generateKustomizeGraph(currentWorkingDirectory, currentWorkingDirectory)

	fmt.Print(graph.String())
}

func generateKustomizeGraph(currentPath string, parent string) {
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}

	kustomizationFilePath, err := findKustomizationFile(currentPath)
	if err != nil {
		log.Fatalf("Could not find kustomization file in path %s", currentPath)
		return
	}

	kustomizationFile, err := readKustomizationFile(kustomizationFilePath)
	if err != nil {
		log.Fatalf("Could not read kustomization file in path %s", currentPath)
		return
	}

	if len(kustomizationFile.Bases) == 0 {
		return
	}

	parentNode := pathToGraphNode(kustomizationFilePath)
	graph.AddNode("main", parentNode, nil)
	for _, base := range kustomizationFile.Bases {
		absoluteBasePath := ""
		if strings.HasPrefix(base, "..") {
			absoluteBasePath, _ = filepath.Abs(path.Join(path.Dir(kustomizationFilePath), base))
		} else {
			absoluteBasePath, _ = filepath.Abs(base)
		}

		childNode := pathToGraphNode(absoluteBasePath)
		if parentNode != childNode {
			graph.AddNode("main", childNode, nil)
			graph.AddEdge(parentNode, childNode, true, nil)
		}

		generateKustomizeGraph(absoluteBasePath, kustomizationFilePath)
	}
}

func pathToGraphNode(path string) string {

	if strings.Contains(path, ".yaml") {
		path, _ = filepath.Split(path)
	}

	path = filepath.Clean(path)
	path = filepath.ToSlash(path)

	path = "\"" + path + "\""

	return path
}

func findKustomizationFile(searchPath string) (string, error) {

	filesInCurrentDirectory, err := ioutil.ReadDir(searchPath)
	if err != nil {
		log.Fatalf("Unable to read directory %s (%s)", searchPath, err.Error())
		return "", err
	}

	foundKustomizeFile := false
	for _, f := range filesInCurrentDirectory {

		if f.IsDir() {
			continue
		}

		if f.Name() == "kustomization.yaml" {
			foundKustomizeFile = true
		}
	}

	if !foundKustomizeFile {
		log.Fatalf("Unable to find kustomization file in path %s", searchPath)
		return "", err
	}

	return path.Join(searchPath, "kustomization.yaml"), nil
}

func readKustomizationFile(kustomizationFilePath string) (KustomizationFileStructure, error) {

	var kustomizationFile KustomizationFileStructure

	readKustomizationFile, err := ioutil.ReadFile(kustomizationFilePath)
	if err != nil {
		log.Fatal("Unable to read kustomization file")
		return kustomizationFile, err
	}

	err = yaml.Unmarshal(readKustomizationFile, &kustomizationFile)
	if err != nil {
		log.Fatal("Unable to parse kustomization file")
		return kustomizationFile, err
	}

	return kustomizationFile, nil
}
