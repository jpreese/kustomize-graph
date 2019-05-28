package kustomizationfile

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

// TestNoKustomizationFiles tests to validate that when no kustomization files
// are found, an error is returned
func TestNoKustomizationFiles(t *testing.T) {
	// Folder structure for this test
	//
	//   /app

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)
	_, err := NewFromFileSystem(fakeFileSystem).GetFromDirectory("app")

	if err == nil {
		t.Errorf("Expected error when reading directory that contains no kustomization files")
	}
}

// TestMultipleKustomizationFiles tests to validate that when multiple kustomization files
// are found, an error is returned
func TestMultipleKustomizationFiles(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   ├── kustomization.yaml
	//   └── kustomization.yml

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)
	emptyFileContents := ""

	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(emptyFileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/kustomization.yml", []byte(emptyFileContents), 0644)
	_, err := NewFromFileSystem(fakeFileSystem).GetFromDirectory("app")

	if err == nil {
		t.Errorf("Expected error when reading directory that contains multiple kustomization files")
	}
}

// TestGetFromDirectory tests the GetFromDirectory method to validate that the kustomization
// yaml file was marshaled correctly from the provided path
func TestGetFromDirectory(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   └── kustomization.yaml

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)

	fileContents := `
resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(fileContents), 0644)
	kustomizationFile, _ := NewFromFileSystem(fakeFileSystem).GetFromDirectory("app")

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

	kustomizationFileContext := NewFromFileSystem(fakeFileSystem)
	kustomizationFile, err := kustomizationFileContext.GetFromDirectory("app")
	if err != nil {
		t.Fatalf("An error occured while getting kustomization file %v", err)
	}
	
	actual, err := kustomizationFileContext.GetMissingResources("app", kustomizationFile)
	if err != nil {
		t.Fatalf("An error occured while getting missing resources %v", err)
	}

	expected := []string{"excluded.yaml"}

	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Returned wrong missing resources, got %s, want: %s", actual, expected)
	}
}
