package main

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// KustomizationFileStructure represents the available attributes in the kustomization yaml file
type KustomizationFileStructure struct {
	Bases     []string `yaml:"bases"`
	Resources []string `yaml:"resources"`
}

var graph = gographviz.NewGraph()

func main() {

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	graphAst, _ := gographviz.ParseString(`digraph main {}`)
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		log.Fatal("Unable to initialize graph")
		return
	}

	err = generateKustomizeGraph(currentWorkingDirectory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(graph.String())
}

func generateKustomizeGraph(currentPath string) error {

	kustomizationFile, err := readKustomizationFile(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	if len(kustomizationFile.Bases) == 0 {
		return nil
	}

	parentNode := sanitizePathForDot(currentPath)
	graph.AddNode("main", parentNode, nil)

	for _, base := range kustomizationFile.Bases {

		absoluteBasePath, _ := filepath.Abs(base)
		if strings.HasPrefix(base, "..") {
			absoluteBasePath, _ = filepath.Abs(path.Join(path.Dir(currentPath), base))
		}

		childNode := sanitizePathForDot(absoluteBasePath)
		graph.AddNode("main", childNode, nil)
		graph.AddEdge(parentNode, childNode, true, nil)

		generateKustomizeGraph(absoluteBasePath)
	}

	return nil
}

func sanitizePathForDot(path string) string {
	path = "\"" + path + "\""
	path = filepath.ToSlash(path)

	return path
}

func readKustomizationFile(kustomizationFilePath string) (KustomizationFileStructure, error) {

	var kustomizationFile KustomizationFileStructure
	kustomizationFilePath = filepath.ToSlash(path.Join(kustomizationFilePath, "kustomization.yaml"))

	readKustomizationFile, err := ioutil.ReadFile(kustomizationFilePath)
	if err != nil {
		return kustomizationFile, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(readKustomizationFile, &kustomizationFile)
	if err != nil {
		return kustomizationFile, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	return kustomizationFile, nil
}
