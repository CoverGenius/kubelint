package main

import (
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
)

/**
* This Linter represents something you can pass in resources to
*	and get results out of that you can eventually log.
*	Also some utility methods for input handling.
**/
type Linter struct {
}

func NewDefaultLinter() *Linter {
	return &Linter{}
}

/**
* Open and lint the files and produce results that
* can be logged later on
**/
func (l *Linter) Lint(filepaths ...string) ([]*Result, error) {
	resources, err := Read(filepaths...)
	if err != nil {
		return nil, err
	}
	var results []*Result
	for _, resource := range resources {
		results = append(results, l.lintResource(resource)...)
	}
	return results, nil
}

func (l *Linter) LintBytes(data []byte, filepath string) []*Result {
	resources := ReadBytes(data, filepath)
	var results []*Result
	for _, resource := range resources {
		results = append(results, l.lintResource(resource)...)
	}
	return results
}

func (l *Linter) LintFile(file *os.File) ([]*Result, error) {
	resources, err := ReadFile(file)
	if err != nil {
		return nil, err
	}
	var results []*Result
	for _, resource := range resources {
		results = append(results, l.lintResource(resource)...)
	}
	return results, nil
}

func (l *Linter) lintResource(resource *YamlDerivedResource) []*Result {
	// figure out the concrete type of this object,
	// then create the rules
	switch resource.Object.(type) {
	case *appsv1.Deployment:
		fmt.Println("This is an appsv1 Deployment!")
	default:
		fmt.Println("Not sure what this is")
	}
	return nil
}
