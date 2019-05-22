package kustomize

import (
	"testing"
	"reflect"
)

func TestGet(t *testing.T) {
	file, _ := NewKustomizationFile().Get("../../testing/get/")

	expected := "a"
	actual := file.Resources[0]

	if actual != expected {
		t.Errorf("Get returned wrong value, got %s, want: %s", actual, expected)
	}
}

func TestGetMissingResources(t *testing.T) {
	file, _ := NewKustomizationFile().Get("../../testing/getmissingresources/")

	actual, _ := file.GetMissingResources()
	expected := []string{"a.yaml"}

	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("GetMissingResources returned wrong resources, got %s, want: %s", actual, expected)
	}
}
