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

	MissingResources []string
}

type kustomizationFileContext struct {
	fileSystem afero.Fs
}

// DefaultContext returns the context to interact with kustomization files
func DefaultContext() *kustomizationFileContext {
	defaultFileSystem := afero.NewOsFs()

	return &kustomizationFileContext{
		fileSystem: defaultFileSystem,
	}
}

// ContextFromFileSystem returns a context based on the given filesystem
func ContextFromFileSystem(fileSystem afero.Fs) *kustomizationFileContext {
	return &kustomizationFileContext{
		fileSystem: fileSystem,
	}
}

// Get attempts to read a kustomization.yaml file
func (k *kustomizationFileContext) Get(filePath string) (*KustomizationFile, error) {
	var kustomizationFile KustomizationFile
	kustomizationFilePath := filepath.ToSlash(path.Join(filePath, "kustomization.yaml"))

	fileUtility := &afero.Afero{Fs: k.fileSystem}
	kustomizationFileBytes, err := fileUtility.ReadFile(kustomizationFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(kustomizationFileBytes, &kustomizationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	missingResources, err := k.getMissingResources(filePath, &kustomizationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get missing resources in path %s", kustomizationFilePath)
	}

	kustomizationFile.MissingResources = missingResources

	return &kustomizationFile, nil
}

func (k *kustomizationFileContext) getMissingResources(filePath string, kustomizationFile *KustomizationFile) ([]string, error) {

	definedResources := []string{}
	definedResources = append(definedResources, kustomizationFile.Resources...)
	definedResources = append(definedResources, kustomizationFile.Patches...)
	definedResources = append(definedResources, kustomizationFile.PatchesStrategicMerge...)

	fileUtility := &afero.Afero{Fs: k.fileSystem}
	directoryInfo, err := fileUtility.ReadDir(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read directory %s", filePath)
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
