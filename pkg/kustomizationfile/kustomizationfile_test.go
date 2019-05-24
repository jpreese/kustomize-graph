package kustomizationfile

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

// TestGet tests the Get method to validate that the kustomization
// yaml file was marshaled correctly from the provided path
func TestGet(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   └── kustomization.yaml
	//

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)

	fileContents := `
resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(fileContents), 0644)
	kustomizationFile, _ := ContextFromFileSystem(fakeFileSystem).Get("app")

	expected := "a.yaml"
	actual := kustomizationFile.Resources[0]

	if actual != expected {
		t.Errorf("Returned wrong resources, got %s, want: %s", actual, expected)
	}
}

// TestGetMissingResources validates that resources that are found in a directory
// but not defined in a kustomization.yaml file are returned correctly
func TestGetMissingResources(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   ├── kustomization.yaml
	//   └── excluded.yaml

	fakeFileSystem := afero.NewMemMapFs()

	fakeFileSystem.Mkdir("app", 0755)
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(""), 0644)
	afero.WriteFile(fakeFileSystem, "app/excluded.yaml", []byte(""), 0644)

	kustomizationFile, err := ContextFromFileSystem(fakeFileSystem).Get("app")
	if err != nil {
		t.Fatalf("An error occured while getting kustomization file %v", err)
	}

	actual := kustomizationFile.MissingResources
	expected := []string{"excluded.yaml"}

	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Returned wrong missing resources, got %s, want: %s", actual, expected)
	}
}
