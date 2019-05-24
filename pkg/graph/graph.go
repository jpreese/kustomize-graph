package graph

import (
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/jpreese/kustomize-graph/pkg/kustomizationfile"
)

// Graph represents the current dependency graph
type Graph interface {
	AddNode(graph string, name string, attributes map[string]string) error
	AddEdge(source string, destination string, directed bool, attributes map[string]string) error
	IsNode(name string) bool
	String() string
}

// KustomizationFileGetter loads an environment to get kustomization files from
type KustomizationFileGetter interface {
	Get(filePath string) (*kustomizationfile.KustomizationFile, error)
}

// NewGraph creates an unpopulated graph with the given name
func NewGraph() *gographviz.Graph {
	graph := gographviz.NewGraph()
	graph.SetName("main")
	graph.Directed = true

	return graph
}

// GenerateKustomizeGraph generates a dependency graph starting from the root path
func GenerateKustomizeGraph(k KustomizationFileGetter, rootPath string) (*gographviz.Graph, error) {

	g := NewGraph()
	err := traverseKustomizeStructure(g, k, rootPath, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Could not produce graph from directory %s", rootPath)
	}

	return g, nil
}

func traverseKustomizeStructure(g Graph, k KustomizationFileGetter, currentPath string, previousNode string) error {
	kustomizationFile, err := k.Get(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not get kustomization file")
	}

	newNode, err := addNodeToGraph(g, currentPath, kustomizationFile)
	if err != nil {
		return errors.Wrapf(err, "Could not create node from path %s", currentPath)
	}

	if previousNode != "" {
		err = g.AddEdge(previousNode, newNode, true, nil)
		if err != nil {
			return errors.Wrapf(err, "Could not create edge from %s to %s", previousNode, newNode)
		}
	}

	// When the kustomization file includes one or more bases
	// we need to recursively call the generateKustomizeGraph method
	// to build out all of the resources present in the base yaml and any
	// other potential bases.
	for _, base := range kustomizationFile.Bases {
		resolveBasePath, err := filepath.Abs(path.Join(currentPath, filepath.Clean(base)))
		if err != nil {
			return errors.Wrapf(err, "Could not resolve base path from base %s and current path %s", base, currentPath)
		}

		err = traverseKustomizeStructure(g, k, resolveBasePath, newNode)
		if err != nil {
			return errors.Wrapf(err, "Error while traversing kustomize structure")
		}
	}

	return nil
}

func addNodeToGraph(g Graph, pathToAdd string, kustomizationFile *kustomizationfile.KustomizationFile) (string, error) {

	node := sanitizePathForDot(pathToAdd)
	if g.IsNode(node) {
		return node, nil
	}

	nodeLabel := getNodeLabelFromMissingResources(pathToAdd, kustomizationFile.MissingResources)

	err := g.AddNode("main", node, nodeLabel)
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
