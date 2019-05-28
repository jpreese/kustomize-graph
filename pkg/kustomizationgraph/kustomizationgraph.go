package kustomizationgraph

import (
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"path"
	"path/filepath"
	"strings"
	"os"

	"github.com/jpreese/kustomize-graph/pkg/kustomizationfile"
)

type kustomizationGraph struct {
	*gographviz.Graph
}

// KustomizationFileGetter gets kustomization files and kustomization file metadata
type KustomizationFileGetter interface {
	GetFromDirectory(directoryPath string) (*kustomizationfile.KustomizationFile, error)
	GetMissingResources(directoryPath string, kustomizationFile *kustomizationfile.KustomizationFile) ([]string, error)
}

// New creates an unpopulated graph with the given name using the given filesystem
func New(graphName string) *kustomizationGraph {
	defaultGraph := gographviz.NewGraph()
	defaultGraph.SetName(graphName)
	defaultGraph.Directed = true

	graph := &kustomizationGraph {
		Graph: defaultGraph,
	}

	return graph
}

// Generate returns a DOT graph based on the dependencies
// from the kustomization.yaml file located in the current working directory
func (g *kustomizationGraph) Generate() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get current working directory")
	}

	kustomizationFileContext := kustomizationfile.New()
	
	err = g.buildGraph(kustomizationFileContext, workingDirectory, "")
	if err != nil {
		return "", errors.Wrapf(err, "Could not produce graph from directory %s", workingDirectory)
	}

	return g.String(), nil
}

func (g *kustomizationGraph) buildGraph(k KustomizationFileGetter, currentPath string, previousNode string) error {
	kustomizationFile, err := k.GetFromDirectory(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not get kustomization file")
	}

	missingResources, err := k.GetMissingResources(currentPath, kustomizationFile)
	if err != nil {
		return errors.Wrapf(err, "Could not get kustomization file missing resources")
	}

	node, err := g.addNodeToGraph(currentPath, missingResources)
	if err != nil {
		return errors.Wrapf(err, "Could not create node from path %s", currentPath)
	}

	if previousNode != "" {
		err = g.AddEdge(previousNode, node, true, nil)
		if err != nil {
			return errors.Wrapf(err, "Could not create edge from %s to %s", previousNode, node)
		}
	}

	// When the kustomization file includes one or more bases we need to recursively call the 
	// buildGraph method to build out all of the resources present in the base yaml and any
	// other potential bases.
	for _, base := range kustomizationFile.Bases {
		resolveBasePath := path.Join(currentPath, filepath.Clean(base))

		err = g.buildGraph(k, resolveBasePath, node)
		if err != nil {
			return errors.Wrapf(err, "Error while traversing kustomize structure")
		}
	}

	return nil
}

func (g *kustomizationGraph) addNodeToGraph(pathToAdd string, missingResources []string) (string, error) {
	node := sanitizePathForDot(pathToAdd)
	if g.IsNode(node) {
		return node, nil
	}

	nodeLabel := getNodeLabel(pathToAdd, missingResources)
	err := g.AddNode(g.Name, node, nodeLabel)
	if err != nil {
		return "", errors.Wrapf(err, "Could not add node %s", node)
	}

	return node, nil
}

func getNodeLabel(filePath string, missingResources []string) map[string]string {
	missingResourcesLabel := make(map[string]string)
	if len(missingResources) == 0 {
		return missingResourcesLabel
	}

	missingPath := filepath.ToSlash(filepath.Clean(filePath))
	nodeLabel := "\"" + missingPath + "\\n\\n"
	nodeLabel += "missing:\\n"
	nodeLabel += strings.Join(missingResources, "\\n")
	nodeLabel += "\""

	missingResourcesLabel["label"] = nodeLabel

	return missingResourcesLabel
}

func sanitizePathForDot(path string) string {
	path = filepath.Clean(path)
	path = "\"" + path + "\""
	path = filepath.ToSlash(path)

	return path
}
