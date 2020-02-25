package main

import (
	"bytes"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

/**
*	Marshals the given resources, using the YAML separator between resources.
**/
func Write(resources ...*Resource) []byte {
	var aggregateBytes []byte
	for i, resource := range resources {
		if o, ok := resource.Object.(runtime.Object); ok {
			bytes, err := stripAndWriteBytes(?, o)
			if err != nil {
				if len(aggregateBytes) != 0 {
					aggregateBytes = append(aggregateBytes, bytes)
				}
			}
		}
	}
	return aggregateBytes
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
