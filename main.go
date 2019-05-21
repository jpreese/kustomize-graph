package main

import
(
	"os"
	"fmt"
	"log"
	"io/ioutil"
	"path/filepath"
	"path"
	"gopkg.in/yaml.v2"
	"github.com/awalterschulze/gographviz"
)

type KustomizationFileStructure struct {
	Bases []string `yaml:"bases"`
	Resources []string `yaml:"resources"`
}

var graphAst, _ = gographviz.ParseString(`digraph main {}`)
var graph = gographviz.NewGraph()

/*
graph.AddNode("main", "a", nil)
graph.AddNode("main", "b", nil)
graph.AddEdge("a", "b", true, nil)
*/

func main() {

	fmt.Println("ok running ...")
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory")
		return
	}

	/*
	rootKustomizationFilePath, err := findKustomizationFile(currentWorkingDirectory);
	if err != nil {
		log.Fatalf("Could not find kustomization file in path %s", rootKustomizationFilePath)
		return
	}

	rootKustomizationFile, err := readKustomizationFile(rootKustomizationFilePath)
	if err != nil {
		log.Fatalf("Could not read kustomization file in path %s", rootKustomizationFilePath)
		return
	}
	*/

	rabbit(currentWorkingDirectory, "")
}

func rabbit(currentPath string, parent string) {
	kustomizationFilePath, err := findKustomizationFile(currentPath);
	if err != nil {
		log.Fatalf("Could not find kustomization file in path %s", currentPath)
		return
	}

	kustomizationFile, err := readKustomizationFile(kustomizationFilePath)
	if err != nil {
		log.Fatalf("Could not read kustomization file in path %s", currentPath)
		return
	}

	graph.AddNode("main", kustomizationFilePath, nil)

	// root node
	if(parent == "") {

	}

	if(len(kustomizationFile.Bases) == 0) {
		return
	}

	for _, base := range kustomizationFile.Bases {
		//absoluteBasePath, _ := filepath.Abs(base)
		//rabbit(absoluteBasePath)
	}
}

func findKustomizationFile(searchPath string) (string, error) {

	filesInCurrentDirectory, err := ioutil.ReadDir(searchPath);
	if err != nil {
		log.Fatalf("Unable to read directory %s (%s)", searchPath, err.Error())
		return "", err
	}

	foundKustomizeFile := false
	for _, f := range filesInCurrentDirectory {

		if(f.IsDir()) {
			continue
		}

		if(f.Name() == "kustomization.yaml") {
			foundKustomizeFile = true
		}
	}

	if(!foundKustomizeFile) {
		log.Fatalf("Unable to find kustomization file in path %s", searchPath)
		return "", err
	}

	fmt.Printf("found a path %s\n", path.Join(searchPath, "kustomization.yaml"))

	return path.Join(searchPath, "kustomization.yaml"), nil
}

func readKustomizationFile(kustomizationFilePath string) (KustomizationFileStructure, error) {

	var kustomizationFile KustomizationFileStructure;

	readKustomizationFile, err := ioutil.ReadFile(kustomizationFilePath);
	if err != nil {
		log.Fatal("Unable to read kustomization file")
		return kustomizationFile, err
	}

	err = yaml.Unmarshal(readKustomizationFile, &kustomizationFile)
	if err != nil {
		log.Fatal("Unable to parse kustomization file")
		return kustomizationFile, err
	}

	return kustomizationFile, nil;
}