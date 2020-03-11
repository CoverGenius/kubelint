package kubelint

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//	This represents a Kubernetes resource.
//	Basically, I don't want to deal with resources that don't conform to these interfaces.
//	It is really difficult to process them and get the necessary information from them.
//	We could always extend this if we get conform them both to the same interface
//	and add in a proxy type for resources not conforming to these interfaces. For now,
//	it'll do.
type Resource struct {
	TypeInfo meta.Type
	Object   metav1.Object
}

//	ConvertToResource attempts to convert any object into a kubernetes Resource.
//	For this to work, the object needs to conform to the v1.Object interface.
//	If it doesn't, and error will be returned.
func ConvertToResource(thing interface{}) (*Resource, error) {
	object, ok := thing.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("Thing could not be converted to a metav1.Object")
	}
	typed, err := meta.TypeAccessor(thing)
	if err != nil {
		return nil, err
	}
	return &Resource{
		TypeInfo: typed,
		Object:   object,
	}, nil
}

//	This is really just a resource,
//	but with some contextual information,
//	so we can have more informative logs.
type YamlDerivedResource struct {
	Resource Resource // the underlying resource

	Filepath   string // the filepath where this resource was found
	LineNumber int    // the line number on which this resource is defined
}
