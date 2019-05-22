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
// and returns the result as a label
type MissingResourceGetter interface {
	GetMissingResources() (map[string]string, error)
}

// GenerateKustomizeGraph generates a dependency graph
func GenerateKustomizeGraph(k KustomizationFileGetter, currentNode string, previousNode string) error {

	kustomizationFile, err := k.Get(currentNode)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentNode)
	}

	newNode, err := addNodeToGraph(kustomizationFile)
	if err != nil {
		return errors.Wrapf(err, "Could not create node from path %s", currentNode)
	}

	if previousNode != "" {
		err = KustomizeGraph.AddEdge(previousNode, newNode, true, nil)
		if err != nil {
			return errors.Wrapf(err, "Could not create edge from %s to %s", previousNode, newNode)
		}
	}

	// When the kustomization file includes one or more bases
	// we need to recursively call the generateKustomizeGraph method
	// to build out all of the resources present in the base yaml and any
	// other potential bases.
	for _, base := range kustomizationFile.Bases {
		absoluteBasePath, _ := filepath.Abs(path.Join(currentNode, strings.TrimPrefix(base, "./")))
		GenerateKustomizeGraph(k, absoluteBasePath, newNode)
	}

	return nil
}

func initializeGraph() *gographviz.Graph {
	graph := gographviz.NewGraph()

	graph.SetName("main")
	graph.Directed = true

	return graph
}

func addNodeToGraph(kustomizationFile KustomizationFile) (string, error) {

	node := sanitizePathForDot(kustomizationFile.Path)
	if KustomizeGraph.IsNode(node) {
		return node, nil
	}

	missingResources, err := kustomizationFile.GetMissingResources()
	if err != nil {
		return "", errors.Wrapf(err, "Could not get missing resources for path %s", kustomizationFile.Path)
	}

	err = KustomizeGraph.AddNode(KustomizeGraph.Name, node, missingResources)
	if err != nil {
		return "", errors.Wrapf(err, "Could not add node %s", node)
	}

	return node, nil
}

func sanitizePathForDot(path string) string {
	path = "\"" + path + "\""
	path = filepath.ToSlash(path)

	return path
}
