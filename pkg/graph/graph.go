package graph

import (
	"github.com/awalterschulze/gographviz"
	"github.com/pkg/errors"
	"path"
	"path/filepath"
	"strings"
	"os"

	"github.com/jpreese/kustomize-graph/pkg/kustomizationfile"
)

// REVIEW NOTE: Method receiver pros/cons with passing around an interface
type kustomizationGraph struct {
	*gographviz.Graph
}

// KustomizationFileGetter loads an environment to get kustomization files from
// REVIEW NOTE: Is it possible to revisit setting the path in the context and use that?
// If going the route of two contexts.. maybe JUST Graph has the notion of Path?
// Adding to this.. could then maybe move missing resources back here.
type KustomizationFileGetter interface {
	Get(filePath string) (*kustomizationfile.KustomizationFile, error)
}

// NewGraph creates an unpopulated graph with the given name
func NewGraph(graphName string) *kustomizationGraph {
	defaultGraph := gographviz.NewGraph()
	defaultGraph.SetName(graphName)
	defaultGraph.Directed = true

	graph := &kustomizationGraph {
		Graph: defaultGraph,
	}

	return graph
}

// GenerateKustomizeGraph returns a DOT graph based on the dependencies
// from the kustomization.yaml file located in the current working directory
func GenerateKustomizeGraph() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get current working directory")
	}

	graph := NewGraph("main")

	// REVIEW NOTE: This guy here. Shouldn't necessarily depend on a kustomization
	// file context, but rather its own context
	kustomizationFileContext := kustomizationfile.DefaultContext()

	err = graph.buildGraph(kustomizationFileContext, workingDirectory, "")
	if err != nil {
		return "", errors.Wrapf(err, "Could not produce graph from directory %s", workingDirectory)
	}

	return graph.String(), nil
}

// REVIEW NOTE: Consider some construct that is responsible for loading all of the
// kustomization files that will be required to build the graph. Then build graph
// will use that construct to actually build it, without needing a filesystem.

// Graph could either have its own filesystem injection OR the packages need to 
// be merged.

// Graph may not need to depend on a filesystem, ultimately just want a collection
// of files... (** though the files themselves contains paths...)

func (g *kustomizationGraph) buildGraph(k KustomizationFileGetter, currentPath string, previousNode string) error {
	kustomizationFile, err := k.Get(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not get kustomization file")
	}

	node, err := g.addNodeToGraph(currentPath, kustomizationFile)
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

func (g *kustomizationGraph) addNodeToGraph(pathToAdd string, kustomizationFile *kustomizationfile.KustomizationFile) (string, error) {
	node := sanitizePathForDot(pathToAdd)
	if g.IsNode(node) {
		return node, nil
	}

	nodeLabel := getNodeLabelFromMissingResources(pathToAdd, kustomizationFile.MissingResources)

	err := g.AddNode(g.Name, node, nodeLabel)
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
