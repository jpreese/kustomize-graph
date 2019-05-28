package kustomizationgraph

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/jpreese/kustomize-graph/pkg/kustomizationfile"
)

// TestGraph tests creating a graph using different ways
// that a base kustomization file can be referenced (e.g.)
//
// - ./singledot
// - justfolder
// - ../updirectory
func TestGraph(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   ├── middle
	//   │   └── kustomization.yaml
	//   ├── same
	//   │   └── kustomization.yaml
	//   ├── base
	//   │   └── kustomization.yaml
	//   └── kustomization.yaml
	//

	fakeFileSystem := afero.NewMemMapFs()

	// Setup 'app' folder
	err := fakeFileSystem.Mkdir("app", 0755)
	appKustomizationFileContents := `
bases:
- same
- ./middle
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(appKustomizationFileContents), 0644)

	// Setup 'same' folder
	fakeFileSystem.Mkdir("app/same", 0755)
	sameKustomizationFileContents := ""
	afero.WriteFile(fakeFileSystem, "app/same/kustomization.yaml", []byte(sameKustomizationFileContents), 0644)

	// Setup 'middle' folder
	fakeFileSystem.Mkdir("app/middle", 0755)
	middleKustomizationFileContents := `
bases:
- ../base
`
	afero.WriteFile(fakeFileSystem, "app/middle/kustomization.yaml", []byte(middleKustomizationFileContents), 0644)

	// Setup 'base' folder
	fakeFileSystem.Mkdir("app/base", 0755)
	baseKustomizationFileContents := ""
	afero.WriteFile(fakeFileSystem, "app/base/kustomization.yaml", []byte(baseKustomizationFileContents), 0644)

	graphContext := NewFromFileSystem(fakeFileSystem, "main")
	kustomizationFileContext := kustomizationfile.NewFromFileSystem(fakeFileSystem)
	
	err = graphContext.buildGraph(kustomizationFileContext, "app", "")
	if err != nil {
		t.Fatalf("Could not generate graph %v.", err)
	}

	// Verify all of the expected nodes are present in the graph
	expectedNodes := []string{
		wrapElement("app"),
		wrapElement("app/same"),
		wrapElement("app/middle"),
		wrapElement("app/base"),
	}
	for _, node := range expectedNodes {
		if !graphContext.IsNode(node) {
			t.Errorf("Expected node %v was not found", node)
		}
	}

	// Verify all of the expected edges are present and their directions are correct
	appFolderEdges := graphContext.Edges.SrcToDsts[wrapElement("app")]
	if _, exists := appFolderEdges[wrapElement("app/middle")]; !exists {
		t.Errorf("Expected edge [app -> app/middle] was not found")
	}
	if _, exists := appFolderEdges[wrapElement("app/same")]; !exists {
		t.Errorf("Expected edge [app -> app/same] was not found")
	}

	middleFolderEdges := graphContext.Edges.SrcToDsts[wrapElement("app/middle")]
	if _, exists := middleFolderEdges[wrapElement("app/base")]; !exists {
		t.Errorf("Expected edge [app/middle -> app/base] was not found")
	}
}

// Elements are stored in the graph with quotes so DOT parses them as strings. 
// This causes problems when attempting to check the their existence, so
// quotes are added here so that they can be found within the returned DOT graph.
func wrapElement(element string) string {
	element = "\"" + element + "\""
	return element
}
