package graph

import (
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/jpreese/kustomize-graph/pkg/kustomize"
)

// KustomizationFileGetter attempts to get information about a kustomization file
type KustomizationFileGetter interface {
	Get(path string) (kustomize.KustomizationFile, error)
	GetMissingResources() ([]string, error)
}

// Graph represents the current dependency graph
type Graph interface {
	AddNode(graph string, name string, attributes map[string]string) error
	AddEdge(source string, destination string, directed bool, attributes map[string]string) error
	IsNode(name string) bool
	String() string
}

// New creates an unpopulated graph with the given name
func New() *gographviz.Graph {
	graph := gographviz.NewGraph()

	graph.SetName("main")
	graph.Directed = true

	return graph
}

// GenerateKustomizeGraph generates a dependency graph
func GenerateKustomizeGraph(g Graph, k kustomize.KustomizationFile) (string, error) {
	err := traverseKustomizeStructure(g, &k, k.Path, "")
	if err != nil {
		return "", errors.Wrapf(err, "Could not produce graph from directory %s", k.Path)
	}

	return g.String(), nil
}

func traverseKustomizeStructure(g Graph, k KustomizationFileGetter, currentPath string, previousNode string) error {
	kustomizationFile, err := k.Get(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	newNode, err := addNodeToGraph(g, &kustomizationFile, currentPath)
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
		absoluteBasePath, _ := filepath.Abs(path.Join(currentPath, strings.TrimPrefix(base, "./")))
		traverseKustomizeStructure(g, k, absoluteBasePath, newNode)
	}

	return nil
}

func addNodeToGraph(g Graph, k KustomizationFileGetter, pathToAdd string) (string, error) {

	node := sanitizePathForDot(pathToAdd)
	if g.IsNode(node) {
		return node, nil
	}

	missingResources, err := k.GetMissingResources()
	if err != nil {
		return "", errors.Wrapf(err, "Could not get missing resources for path %s", pathToAdd)
	}

	nodeLabel := getNodeLabelFromMissingResources(pathToAdd, missingResources)

	err = g.AddNode("main", node, nodeLabel)
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
