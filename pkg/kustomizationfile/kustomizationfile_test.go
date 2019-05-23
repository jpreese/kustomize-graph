package kustomizationfile

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func NewFake(fakeFileSystem afero.Fs, path string) *kustomizationFileLoader {
	return &kustomizationFileLoader{
		fileSystem: fakeFileSystem,
		path:       path,
	}
}

// Folder structure for this test
//
//   /app
//   └── kustomization.yaml
//
func TestGet(t *testing.T) {
	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)

	fileContents := `
resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(fileContents), 0644)
	kustomizationFile, _ := NewFake(fakeFileSystem, "app").Get()

	expected := "a.yaml"
	actual := kustomizationFile.Resources[0]

	if actual != expected {
		t.Fatalf("Returned wrong resources, got %s, want: %s", actual, expected)
	}
}

// Folder structure for this test
//
//   /app
//   ├── kustomization.yaml
//   └── excluded.yaml
func TestGetMissingResources(t *testing.T) {
	fakeFileSystem := afero.NewMemMapFs()

	rootPath := "app"

	fakeFileSystem.Mkdir(rootPath, 0755)
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(""), 0644)
	afero.WriteFile(fakeFileSystem, "app/excluded.yaml", []byte(""), 0644)

	actual, _ := NewFake(fakeFileSystem, rootPath).GetMissingResources()
	expected := []string{"excluded.yaml"}

	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Rreturned wrong missing resources, got %s, want: %s", actual, expected)
	}
}
