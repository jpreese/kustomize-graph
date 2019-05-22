package main

import (
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"path"
	"path/filepath"
	"strings"
)

// KustomizeGraph represents the kustomize dependency graph
var KustomizeGraph = initializeGraph()

// KustomizationFileGetter attempts to get a kustomization file
type KustomizationFileGetter interface {
	Get(path string) (KustomizationFile, error)
}

// MissingResourceGetter gets all of the resources missing from a kustomization file
type MissingResourceGetter interface {
	GetMissingResources() ([]string, error)
}

// GenerateKustomizeGraph generates a dependency graph
func GenerateKustomizeGraph(k KustomizationFileGetter, currentPath string, previousNode string) (*gographviz.Graph, error) {

	kustomizationFile, err := k.Get(currentPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	newNode, err := addNodeToGraph(&kustomizationFile, currentPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create node from path %s", currentPath)
	}

	if previousNode != "" {
		err = KustomizeGraph.AddEdge(previousNode, newNode, true, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not create edge from %s to %s", previousNode, newNode)
		}
	}

	// When the kustomization file includes one or more bases
	// we need to recursively call the generateKustomizeGraph method
	// to build out all of the resources present in the base yaml and any
	// other potential bases.
	for _, base := range kustomizationFile.Bases {
		absoluteBasePath, _ := filepath.Abs(path.Join(currentPath, strings.TrimPrefix(base, "./")))
		GenerateKustomizeGraph(k, absoluteBasePath, newNode)
	}

	return KustomizeGraph, nil
}

func initializeGraph() *gographviz.Graph {
	graph := gographviz.NewGraph()

	graph.SetName("main")
	graph.Directed = true

	return graph
}

func addNodeToGraph(m MissingResourceGetter, pathToAdd string) (string, error) {

	node := sanitizePathForDot(pathToAdd)
	if KustomizeGraph.IsNode(node) {
		return node, nil
	}

	missingResources, err := m.GetMissingResources()
	if err != nil {
		return "", errors.Wrapf(err, "Could not get missing resources for path %s", pathToAdd)
	}

	nodeLabel := getNodeLabelFromMissingResources(pathToAdd, missingResources)

	err = KustomizeGraph.AddNode(KustomizeGraph.Name, node, nodeLabel)
	if err != nil {
		return "", errors.Wrapf(err, "Could not add node %s", node)
	}

	return node, nil
}

func getNodeLabelFromMissingResources(filePath string, missingResources []string) map[string]string {

	missingResourcesLabel := make(map[string]string)
	if len(missingResources) == 0 {
		return missingResourcesLabel
	}

	nodeLabel := "\"" + filepath.ToSlash(filePath) + "\\n\\n"
	nodeLabel += "missing:\\n"
	nodeLabel += strings.Join(missingResources, "\\n")
	nodeLabel += "\""

	missingResourcesLabel["label"] = nodeLabel

	return missingResourcesLabel
}

func sanitizePathForDot(path string) string {
	path = "\"" + path + "\""
	path = filepath.ToSlash(path)

	return path
}
