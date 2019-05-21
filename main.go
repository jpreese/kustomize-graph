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
	Bases   []string `yaml:"bases"`
	Patches []string `yaml:"patches"`
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

	parent := sanitizePathForDot(currentPath)
	graph.AddNode("main", parent, nil)

	for _, base := range kustomizationFile.Bases {
		handleBase(parent, base, currentPath)
	}

	return nil
}

func handleBase(parent string, value string, context string) {
	absoluteBasePath, _ := filepath.Abs(value)
	if strings.HasPrefix(value, "..") {
		absoluteBasePath, _ = filepath.Abs(path.Join(path.Dir(context), value))
	}

	child := sanitizePathForDot(absoluteBasePath)
	graph.AddNode("main", child, nil)
	graph.AddEdge(parent, child, true, nil)

	generateKustomizeGraph(absoluteBasePath)
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
