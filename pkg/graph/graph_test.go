package graph

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/jpreese/kustomize-graph/pkg/kustomizationfile"
)

// TestGraph tests creating a graph using different ways 
// that a base kustomization file can be referenced:
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

	fakeFileSystem := afero.NewBasePathFs(afero.NewMemMapFs(), "/app/same/lol")

	// Setup app folder kustomization file
	err := fakeFileSystem.MkdirAll("/tmp/app", 0755)
	appKustomizationFileContents := `
bases:
- same
- ./middle
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(appKustomizationFileContents), 0644)

	// Setup same folder kustomization file
	fakeFileSystem.Mkdir("app/same", 0755)
	sameKustomizationFileContents := ""
	afero.WriteFile(fakeFileSystem, "/app/same/kustomization.yaml", []byte(sameKustomizationFileContents), 0644)

	// Setup middle folder kustomization file
	fakeFileSystem.Mkdir("app/middle", 0755)
	middleKustomizationFileContents := `
bases:
- ../base
`
	afero.WriteFile(fakeFileSystem, "/app/middle/kustomization.yaml", []byte(middleKustomizationFileContents), 0644)

	// Setup base folder kustomization file
	fakeFileSystem.Mkdir("app/base", 0755)
	baseKustomizationFileContents := ""
	afero.WriteFile(fakeFileSystem, "/app/base/kustomization.yaml", []byte(baseKustomizationFileContents), 0644)

	graph, err := GenerateKustomizeGraph(kustomizationfile.ContextFromFileSystem(fakeFileSystem), "/app")
	if err != nil {
		t.Fatalf("Could not generate graph %v", err)
	}

	// Verify all of the expected nodes are present in the graph
	expectedNodes := []string{ 
		wrapElement("app"),
		wrapElement("app/same"),
		wrapElement("app/middle"),
		wrapElement("app/base"),
	}
	for _, node := range expectedNodes {
		if(!graph.IsNode(node)) {
			t.Fatalf("Expected node %v was not found. Nodes are %v. FileSystem was %v", node, graph.Nodes, fakeFileSystem)
		}
	}

	// Verify all of the expected edges are present
	appEdges := graph.Edges.SrcToDsts[wrapElement("app")]
	if _, exists := appEdges[wrapElement("app/middle")]; !exists {
		t.Fatalf("Expected edge [app -> app/middle] was not found")
	}
	if _, exists := appEdges[wrapElement("app/same")]; !exists {
		t.Fatalf("Expected edge [app -> app/same] was not found")
	}

	middleEdges := graph.Edges.SrcToDsts[wrapElement("app/middle")]
	if _, exists := middleEdges[wrapElement("app/base")]; !exists {
		t.Fatalf("Expected edge [app/middle -> app/base] was not found")
	}
}

// Elements are stored in the graph with quotes so DOT
// parses them as strings. Though this causes problems
// when attempting to check the their existence
func wrapElement(element string) string {
	element = "\"" + element + "\""
	return element
}
