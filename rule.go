package main

import (
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Rule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func() bool
	Message        string
	Level          log.Level
	Resources      []*YamlDerivedResource
	Fix            func() bool
	FixDescription func() string
}

// The unique identifier for a rule. This lets us define an execution order with the Prereqs field.
type RuleID string

/**
*	This represents a generic rule that can be applied to a deployment object.
* 	All other <Resource>Rule structs are analogous.
**/
type DeploymentRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*appsv1.Deployment) bool
	Message        string
	Level          log.Level
	Fix            func(*appsv1.Deployment) bool
	FixDescription func(*appsv1.Deployment) string
}

/**
*	Once we get a reference to an actual resource, we can interpolate this into the
*	method bodies, and let every rule conform to the same structure.
*	At this point, we have no information about where this resource came from.
**/
func (d *DeploymentRule) CreateRule(deployment *appsv1.Deployment) *Rule {
	var typeErased interface{} = deployment
	resource := &YamlDerivedResource{
		Resource: Resource{TypeInfo: typeErased.(meta.Type), Object: typeErased.(metav1.Object)},
	}
	r := &Rule{
		ID:             d.ID,
		Prereqs:        d.Prereqs,
		Condition:      func() bool { return d.Condition(deployment) },
		Message:        d.Message,
		Level:          d.Level,
		Resources:      []*YamlDerivedResource{resource},
		Fix:            func() bool { return d.Fix(deployment) },
		FixDescription: func() string { return d.FixDescription(deployment) },
	}
	return r
}
