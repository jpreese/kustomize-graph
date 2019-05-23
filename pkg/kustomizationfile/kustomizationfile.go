package kustomizationfile

import (
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// KustomizationFile represents a kustomization yaml file
type KustomizationFile struct {
	Bases                 []string `yaml:"bases"`
	Resources             []string `yaml:"resources"`
	Patches               []string `yaml:"patches"`
	PatchesStrategicMerge []string `yaml:"patchesStrategicMerge"`
}

type loader struct {
	fileSystem afero.Fs
}

// New creates a loader to get kustomization files from
func New() *loader {
	defaultFileSystem := afero.NewOsFs()

	return &loader{
		fileSystem: defaultFileSystem,
	}
}

// Get attempts to read a kustomization.yaml file
func (l *loader) Get() (*KustomizationFile, error) {
	var kustomizationFile KustomizationFile
	kustomizationFilePath := filepath.ToSlash(path.Join(loader.Path, "kustomization.yaml"))

	fileUtility := &afero.Afero{Fs: loader.fileSystem}
	kustomizationFileBytes, err := fileUtility.ReadFile(kustomizationFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(kustomizationFileBytes, &kustomizationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	return &kustomizationFile, nil
}

// GetMissingResources finds all of the resources that exist in the folder
// but are not explicitly defined in the kustomization.yaml file
func (loader *Loader) GetMissingResources() ([]string, error) {

	kustomizationFile, err := loader.Get()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not load file from path %s", loader.Path)
	}

	definedResources := []string{}
	definedResources = append(definedResources, kustomizationFile.Resources...)
	definedResources = append(definedResources, kustomizationFile.Patches...)
	definedResources = append(definedResources, kustomizationFile.PatchesStrategicMerge...)

	fileUtility := &afero.Afero{Fs: loader.FileSystem}
	directoryInfo, err := fileUtility.ReadDir(loader.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read directory %s", loader.Path)
	}

	missingResources := []string{}
	for _, info := range directoryInfo {

		if info.IsDir() {
			continue
		}

		// Only consider the resource missing if it is a yaml file
		if filepath.Ext(info.Name()) != ".yaml" {
			continue
		}

		// Ignore the kustomization file itself
		if info.Name() == "kustomization.yaml" {
			continue
		}

		if !existsInSlice(definedResources, info.Name()) {
			missingResources = append(missingResources, info.Name())
		}
	}

	return missingResources, nil
}

func existsInSlice(slice []string, element string) bool {
	for _, current := range slice {
		if current == element {
			return true
		}
	}
	return false
}
