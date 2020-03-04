package tests

import (
	"testing"

	"github.com/CoverGenius/kubelint"
)

func TestReadBytes(t *testing.T) {
	deploymentDefinition := []byte(`
kind: Deployment
apiVersion: apps/v1
metadata:
  name: hello-world
`)
	resources, errs := kubelint.ReadBytes(deploymentDefinition, "FAKE_FILE.yaml")
	for _, err := range errs {
		t.Error(err)
	}
	for _, resource := range resources {
		t.Logf("%#v\n", resource.Resource)
	}
	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource to be successfully parsed")
	}
	if resources[0].Resource.Object.GetName() != "hello-world" {
		t.Errorf("Resource name incorrectly set")
	}
}
