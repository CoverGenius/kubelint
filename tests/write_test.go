package tests

import (
	"fmt"
	"testing"

	"github.com/rdowavic/kubelint"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWriteOneResource(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
	}
	resource, err := ConvertToResource(deployment)
	if err != nil {
		t.Fatal(err)
	}
	bytes, errors := kubelint.Write(resource)
	t.Log(string(bytes))
	for _, err := range errors {
		t.Error(err)
	}
}

func ConvertToResource(thing interface{}) (*kubelint.Resource, error) {
	object, ok := thing.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("Thing could not be converted to a metav1.Object")
	}
	typed, err := meta.TypeAccessor(thing)
	if err != nil {
		return nil, err
	}
	return &kubelint.Resource{
		TypeInfo: typed,
		Object:   object,
	}, nil
}

func TestWriteTwoResources(t *testing.T) {
	r1, err := ConvertToResource(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "my-first-deployment"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := ConvertToResource(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "my-second-deployment"},
	})
	if err != nil {
		t.Fatal(err)
	}
	bytes, errors := kubelint.Write(r1, r2)
	t.Log(string(bytes))
	for _, err := range errors {
		t.Error(err)
	}
}
