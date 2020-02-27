package kubelint

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

/**
*	Writes the passed resources to the given file in their YAML representation
*	Eg. Resource{*appsv1.Deployment{..}} -> apiVersion: apps/v1\nkind:Deployment\n ...
**/
func WriteToFile(file *os.File, resources ...*Resource) []error {
	bytes, errors := Write(resources...)
	_, err := file.Write(bytes)
	errors = append(errors, err)
	return errors
}

/**
*	Marshals the given resources, using the YAML separator between resources.
**/
func Write(resources ...*Resource) ([]byte, []error) {
	var aggregateBytes []byte
	var errors []error
	// serialiser tool
	serialiser := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	documentSeparator := []byte("---\n")

	for _, resource := range resources {
		if o, ok := resource.Object.(runtime.Object); ok {
			bytes, err := stripAndWriteBytes(serialiser, o)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			if len(aggregateBytes) != 0 {
				// then you should write up a separator
				aggregateBytes = append(aggregateBytes, documentSeparator...)
			}
			aggregateBytes = append(aggregateBytes, bytes...)
		} else {
			errors = append(errors, fmt.Errorf("Resource could not be serialised because it doesn't conform to the runtime.Object interface"))
		}
	}
	return aggregateBytes, errors
}

func stripAndWriteBytes(s *json.Serializer, o runtime.Object) ([]byte, error) {
	var b bytes.Buffer
	err := s.Encode(o, &b)
	if err != nil {
		return nil, err
	}
	str := b.String()
	re := regexp.MustCompile(" *creationTimestamp: null\n")
	cleaned := re.ReplaceAllString(str, "")
	return []byte(cleaned), nil
}
