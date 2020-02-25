package main

import (
	log "github.com/sirupsen/logrus"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/**
*	This represents a Kubernetes resource.
*	Basically, I don't want to deal with resources that don't conform to these interfaces.
*	It is really difficult to process them and get the necessary information from them.
*	We could always extend this if we get conform them both to the same interface
*	and add in a proxy type for resources not conforming to these interfaces. For now,
*	it'll do.
**/
type Resource struct {
	TypeInfo meta.Type
	Object   metav1.Object
}

/**
*	This is really just a resource,
*	but with some contextual information,
*	so we can have more informative logs.
**/
type YamlDerivedResource struct {
	Resource

	Filepath   string // the filepath where this resource was found
	LineNumber int    // the line number on which this resource is defined
}

/**
*	Struct to carry all information necessary for the logger.
**/
type Result struct {
	Resource *YamlDerivedResource // the resource on which the rule was performed to get this result
	Message  string               // the complaining message (eg "no securityContextKey present")
	Level    log.Level            // the level of trouble this result causes
}
