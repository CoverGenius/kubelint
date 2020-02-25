package main

import (
	bytesPkg "bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

/**
*	Given a list of filenames to read from, produce
*	YamlDerivedResources
**/
func Read(filepaths ...string) ([]*YamlDerivedResource, error) {
	var resources []*YamlDerivedResource
	for _, filepath := range filepaths {
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		resources = append(resources, ReadBytes(content, filepath)...)
	}
	return resources, nil
}

func ReadFile(file *os.File) ([]*YamlDerivedResource, error) {
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	resources := ReadBytes(content, file.Name())
	return resources, nil
}

/**
* ReadBytes takes in a bytes representation of a kubernetes object declaration
* and attempts to construct the concrete in-memory representation of them.
* It will silently fail if something doesn't conform to the Resource struct requirements (meta.Type and metav1.Object conformance)
* I may have to change this in the future.
**/
func ReadBytes(bytes []byte, filepath string) []*YamlDerivedResource {
	var resources []*YamlDerivedResource
	newline := detectLineBreak(bytes)
	segments := bytesPkg.Split(bytes, []byte(fmt.Sprintf("%s---%s", newline, newline)))
	lineNumber := findLineNumbers(bytes)
	currentObjNum := 0
	// 1. Iterate over each byte representation of an object
	for _, marshalledResource := range segments {
		if len(strings.Trim(string(marshalledResource), newline)) == 0 {
			continue
		}
		// 2. Decode the object into its corresponding k8s type (eg *appsv1.Deployment)
		concrete, _, err := scheme.Codecs.UniversalDeserializer().Decode(marshalledResource, nil, nil)
		if err != nil {
			// possibly should log something here
			continue
		}
		// 3. Try to get the object to conform to these easy-to-use interfaces
		typeInfo, ok := concrete.(meta.Type)
		if !ok {
			// possibly should log something here
			continue
		}
		object, ok := concrete.(metav1.Object)
		if !ok {
			// possibly should log something here
			continue
		}
		resources = append(resources, &YamlDerivedResource{
			Filepath:   filepath,
			LineNumber: lineNumber[currentObjNum],
			Resource: Resource{
				TypeInfo: typeInfo,
				Object:   object,
			},
		})
		currentObjNum++
	}
	return resources
}

// copied from https://github.com/instrumenta/kubeval/blob/9c9c0a5b3cc619dbd94129af77c8512bfd0f1763/kubeval/utils.go#L24
func detectLineBreak(haystack []byte) string {
	windowsLineEnding := bytesPkg.Contains(haystack, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

/**
* For each object (in the order that they occur in the yaml file), tell me what line number the object starts on.
* This is brittle, will break as soon as kubernetes objects aren't given the apiVersion as the first key sorry about this.
 */
func findLineNumbers(data []byte) []int {
	objectSignifier := []byte("apiVersion:")
	numObjects := bytesPkg.Count(data, objectSignifier)
	lineNum := make([]int, numObjects)
	currentObject := 0
	newline := []byte(detectLineBreak(data))
	for i, line := range bytesPkg.Split(data, newline) {
		if bytesPkg.Contains(line, objectSignifier) {
			lineNum[currentObject] = i + 1
			currentObject += 1
		}
	}
	return lineNum
}
