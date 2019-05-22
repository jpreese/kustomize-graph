package kustomizegraph

import (
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"path"
	"path/filepath"
	"strings"
)

var GraphName = "main"
var KustomizeGraph = gographviz.NewGraph()

func GenerateKustomizeGraph(currentPath string, previousNode string) error {

	KustomizeGraph.SetName(GraphName)
	KustomizeGraph.Directed = true

	kustomizationFile, err := NewKustomizationFile().GetKustomizationFile(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	node, err := addNodeToGraph(kustomizationFile)
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

		GenerateKustomizeGraph(absoluteBasePath, node)
	}

	return nil
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

	err = KustomizeGraph.AddNode(GraphName, node, missingResources)
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
