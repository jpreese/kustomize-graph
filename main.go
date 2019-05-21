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
	Resources []string `yaml:"resources"`
	Patches []string `yaml:"patches"`
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

	err = generateKustomizeGraph(currentWorkingDirectory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(KustomizeGraph.String())
}

func generateKustomizeGraph(currentPath string) error {
	kustomizationFile, err := readKustomizationFile(currentPath)
	if err != nil {
		return errors.Wrapf(err, "Could not read kustomization file in path %s", currentPath)
	}

	parentNode := sanitizePathForDot(currentPath)
	KustomizeGraph.AddNode("main", parentNode, nil)

	// Handle the resources present in the kustomization file that 
	// are not in the bases section. This typically includes:
	// resources and patches where recursion isn't needed.
	handleRelativeResources(kustomizationFile, currentPath, parentNode)
	
	// When the kustomization file includes one or more bases
	// we need to recursively call the generateKustomizeGraph method
	// to build out all of the resources present in the base yaml and any
	// other potential additional bases.
	for _, base := range kustomizationFile.Bases {
		absoluteBasePath, _ := filepath.Abs(base)
		if strings.HasPrefix(base, "..") {
			absoluteBasePath, _ = filepath.Abs(path.Join(path.Dir(currentPath), base))
		}
	
		childNode := sanitizePathForDot(absoluteBasePath)
		addChildNodeToParent(childNode, parentNode)
	
		// Recursively call this method again to resolve additional bases
		generateKustomizeGraph(absoluteBasePath)
	}

	return nil
}

func handleRelativeResources(kustomizationFile KustomizationFileStructure, currentPath string, parent string) {
	for _, resource := range kustomizationFile.Resources {
		resourceNode := sanitizePathForDot(joinFileNameToPath(currentPath, resource))
		addChildNodeToParent(resourceNode, parent)
	}

	for _, patch := range kustomizationFile.Patches {
		patchNode := sanitizePathForDot(joinFileNameToPath(currentPath, patch))
		addChildNodeToParent(patchNode, parent)
	}

	for _, patchStrategicMerge := range kustomizationFile.PatchesStrategicMerge {
		patchStrategicMergeNode := sanitizePathForDot(joinFileNameToPath(currentPath, patchStrategicMerge))
		addChildNodeToParent(patchStrategicMergeNode, parent)
	}
}

func addChildNodeToParent(childNode string, parentNode string) {
	KustomizeGraph.AddNode("main", childNode, nil)
	KustomizeGraph.AddEdge(parentNode, childNode, true, nil)
}

func sanitizePathForDot(path string) string {
	path = "\"" + path + "\""
	path = filepath.ToSlash(path)

	return path
}

func readKustomizationFile(kustomizationFilePath string) (KustomizationFileStructure, error) {
	var kustomizationFile KustomizationFileStructure
	kustomizationFilePath = joinFileNameToPath(kustomizationFilePath, "kustomization.yaml")

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

func joinFileNameToPath(filePath string, fileName string) string {
	return filepath.ToSlash(path.Join(filePath, fileName))
}
