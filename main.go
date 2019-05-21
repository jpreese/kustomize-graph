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

var graphAst, _ = gographviz.ParseString(`digraph main {}`)
var graph = gographviz.NewGraph()

func main() {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	err = gographviz.Analyse(graphAst, graph)
	if err != nil {
		log.Fatal("Unable to initialize dependency graph")
		return
	}

	err = generateKustomizeGraph(currentWorkingDirectory, currentWorkingDirectory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(graph.String())
}

func generateKustomizeGraph(currentPath string, parent string) error {

	kustomizationFilePath, err := findKustomizationFile(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not find kustomization file in path %s", currentPath)
	}

	kustomizationFile, err := readKustomizationFile(kustomizationFilePath)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	if len(kustomizationFile.Bases) == 0 {
		return nil
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

	return nil
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
		return "", errors.Wrapf(err, "Unable to read directory %s", searchPath)
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
		return "", errors.Wrapf(err, "Unable to find kustomization file in path %s", searchPath)
	}

	return path.Join(searchPath, "kustomization.yaml"), nil
}

func readKustomizationFile(kustomizationFilePath string) (KustomizationFileStructure, error) {

	var kustomizationFile KustomizationFileStructure

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
