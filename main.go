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
	Bases                 []string `yaml:"bases"`
	Resources             []string `yaml:"resources"`
	Patches               []string `yaml:"patches"`
	PatchesStrategicMerge []string `yaml:"patchesStrategicMerge"`
}

// KustomizeGraph represents the generated DOT graph
var KustomizeGraph = gographviz.NewGraph()

func main() {

	graphAst, _ := gographviz.ParseString(`digraph main {}`)
	if err := gographviz.Analyse(graphAst, KustomizeGraph); err != nil {
		log.Fatal("Unable to initialize graph")
		return
	}

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	err = generateKustomizeGraph(currentWorkingDirectory, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(KustomizeGraph.String())
}

func generateKustomizeGraph(currentPath string, previousNode string) error {
	kustomizationFile, err := readKustomizationFile(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	node, err := addNodeToGraph(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not create node from path %s", currentPath)
	}

	if previousNode != "" {
		err = KustomizeGraph.AddEdge(previousNode, node, true, nil)
		if err != nil {
			return errors.Wrapf(err, "Could not create edge from %s to %s", previousNode, node)
		}
	}

	// When the kustomization file includes one or more bases
	// we need to recursively call the generateKustomizeGraph method
	// to build out all of the resources present in the base yaml and any
	// other potential bases.
	for _, base := range kustomizationFile.Bases {
		absoluteBasePath, _ := filepath.Abs(path.Join(currentPath, strings.TrimPrefix(base, "./")))

		generateKustomizeGraph(absoluteBasePath, node)
	}

	return nil
}

func addNodeToGraph(path string) (string, error) {

	node := sanitizePathForDot(path)
	if KustomizeGraph.IsNode(node) {
		return node, nil
	}

	kustomizationFile, err := readKustomizationFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "Could not read kustomization file in path %s", path)
	}

	missingResources, err := getMissingResourceAttributes(kustomizationFile, path)
	if err != nil {
		return "", errors.Wrapf(err, "Could not get excluded resource attributes for path %s", path)
	}

	err = KustomizeGraph.AddNode("main", node, missingResources)
	if err != nil {
		return "", errors.Wrapf(err, "Could not add node %s", node)
	}

	return node, nil
}

func getMissingResourceAttributes(kustomizationFile KustomizationFileStructure, currentPath string) (map[string]string, error) {

	definedYamls := []string{}
	definedYamls = append(definedYamls, kustomizationFile.Resources...)
	definedYamls = append(definedYamls, kustomizationFile.Patches...)
	definedYamls = append(definedYamls, kustomizationFile.PatchesStrategicMerge...)

	foundMissingResources, err := findMissingResources(currentPath, definedYamls)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find missing resources (%s) in path %s", definedYamls, currentPath)
	}

	nodeAttributes := make(map[string]string)

	if len(foundMissingResources) == 0 {
		return nodeAttributes, nil
	}

	nodeAttributes["label"] = strings.Join(foundMissingResources, ",")
	return nodeAttributes, nil
}

func findMissingResources(pathToSearch string, filesToCheck []string) ([]string, error) {

	directoryInfo, err := ioutil.ReadDir(pathToSearch)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read directory %s", pathToSearch)
	}

	missingFiles := []string{}
	for _, info := range directoryInfo {

		if info.IsDir() {
			continue
		}

		// Only consider the resource missing if it is a yaml file
		if filepath.Ext(info.Name()) != ".yaml" {
			continue
		}

		// Ignore the kustomization file itself
		if info.Name() == "kustomization.yaml" {
			continue
		}

		if !existsInSlice(filesToCheck, info.Name()) {
			missingFiles = append(missingFiles, info.Name())
		}
	}

	return missingFiles, nil
}

func existsInSlice(slice []string, element string) bool {
	for _, current := range slice {
		if current == element {
			return true
		}
	}
	return false
}

func sanitizePathForDot(path string) string {
	path = "\"" + path + "\""
	path = filepath.ToSlash(path)

	return path
}

func readKustomizationFile(kustomizationFilePath string) (KustomizationFileStructure, error) {

	var kustomizationFile KustomizationFileStructure

	kustomizationFilePath = filepath.ToSlash(path.Join(kustomizationFilePath, "kustomization.yaml"))

	kustomizationFileBytes, err := ioutil.ReadFile(kustomizationFilePath)
	if err != nil {
		return kustomizationFile, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(kustomizationFileBytes, &kustomizationFile)
	if err != nil {
		return kustomizationFile, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	return kustomizationFile, nil
}
