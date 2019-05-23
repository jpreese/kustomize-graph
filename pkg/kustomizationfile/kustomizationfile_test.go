package kustomizationfile

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func FakeContext(fakeFileSystem afero.Fs) *kustomizationFileContext {
	return &kustomizationFileContext{
		fileSystem: fakeFileSystem,
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
	kustomizationFile, _ := FakeContext(fakeFileSystem).Get("app")

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

	kustomizationFile, _ := FakeContext(fakeFileSystem).Get("app")

	actual := kustomizationFile.MissingResources
	expected := []string{"excluded.yaml"}

	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Returned wrong missing resources, got %s, want: %s", actual, expected)
	}
}
