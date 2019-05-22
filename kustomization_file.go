package kustomizegraph

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

// KustomizationFile represents the kustomization.yaml file
type KustomizationFile struct {
	Path                  string
	Bases                 []string `yaml:"bases"`
	Resources             []string `yaml:"resources"`
	Patches               []string `yaml:"patches"`
	PatchesStrategicMerge []string `yaml:"patchesStrategicMerge"`
}

// NewKustomizationFile creates an empty kustomization representation
func NewKustomizationFile() *KustomizationFile {
	return &KustomizationFile{}
}

// GetKustomizationFile attempts to read a kustomization.yaml file from the specified path
func (file *KustomizationFile) GetKustomizationFile(filePath string) (KustomizationFile, error) {

	var kustomizationFile KustomizationFile

	kustomizationFilePath := filepath.ToSlash(path.Join(filePath, "kustomization.yaml"))

	kustomizationFileBytes, err := ioutil.ReadFile(kustomizationFilePath)
	if err != nil {
		return kustomizationFile, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(kustomizationFileBytes, &kustomizationFile)
	if err != nil {
		return kustomizationFile, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	kustomizationFile.Path = filePath

	return kustomizationFile, nil
}

// GetMissingResources finds all of the resources that exist in the folder
// but are not explicitly defined in the kustomization.yaml file
func (file *KustomizationFile) GetMissingResources() (map[string]string, error) {

	definedResources := []string{}
	definedResources = append(definedResources, file.Resources...)
	definedResources = append(definedResources, file.Patches...)
	definedResources = append(definedResources, file.PatchesStrategicMerge...)

	directoryInfo, err := ioutil.ReadDir(file.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read directory %s", file.Path)
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

	if len(missingResources) == 0 {
		return map[string]string{}, nil
	}

	missingResourcesLabel := make(map[string]string)
	missingResourcesLabel["label"] = getMissingResourceLabel(file.Path, missingResources)

	return missingResourcesLabel, nil
}

func getMissingResourceLabel(path string, missingResources []string) string {
	label := "\"" + filepath.ToSlash(path) + "\\n\\n"
	label += "missing:\\n"
	label += strings.Join(missingResources, "\\n")
	label += "\""

	return label
}

func existsInSlice(slice []string, element string) bool {
	for _, current := range slice {
		if current == element {
			return true
		}
	}
	return false
}
