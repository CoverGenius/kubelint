package kubelint

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	batchV1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	v1beta1Extensions "k8s.io/api/extensions/v1beta1"
	networkingV1 "k8s.io/api/networking/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	rbacV1beta1 "k8s.io/api/rbac/v1beta1"

	"os"
)

/**
* Linter: This Linter represents something you can pass in resources to
*	and get results out of that you can eventually log.
*	Also some utility methods for input handling.
**/
type Linter struct {
	appsV1DeploymentRules               []*AppsV1DeploymentRule               // a register for all user-defined appsV1Deployment rules
	v1NamespaceRules                    []*V1NamespaceRule                    // a register for all user-defined v1Namespace rules
	v1PersistentVolumeClaimRules        []*V1PersistentVolumeClaimRule        // a register for all user-defined v1PersistentVolumeClaim rules
	v1Beta1ExtensionsDeploymentRules    []*V1Beta1ExtensionsDeploymentRule    // a register for all user-defined v1Beta1ExtensionsDeployment rules
	batchV1JobRules                     []*BatchV1JobRule                     // a register for all user-defined batchV1Job rules
	batchV1Beta1CronJobRules            []*BatchV1Beta1CronJobRule            // a register for all user-defined batchV1Beta1CronJob rules
	v1Beta1ExtensionsIngressRules       []*V1Beta1ExtensionsIngressRule       // a register for all user-defined v1Beta1ExtensionsIngress rules
	networkingV1NetworkPolicyRules      []*NetworkingV1NetworkPolicyRule      // a register for all user-defined networkingV1NetworkPolicy rules
	v1Beta1ExtensionsNetworkPolicyRules []*V1Beta1ExtensionsNetworkPolicyRule // a register for all user-defined v1Beta1ExtensionsNetworkPolicy rules
	rbacV1RoleRules                     []*RbacV1RoleRule                     // a register for all user-defined rbacV1Role rules
	rbacV1Beta1RoleBindingRules         []*RbacV1Beta1RoleBindingRule         // a register for all user-defined rbacV1Beta1RoleBinding rules
	v1ServiceAccountRules               []*V1ServiceAccountRule               // a register for all user-defined v1ServiceAccount rules
	v1ServiceRules                      []*V1ServiceRule                      // a register for all user-defined v1Service rules
	fixes                               []*RuleSorter                         // fixes that should be applied to the resources in order to mitigate some errors on a future pass
	resources                           []*Resource                           // All the resources that have been read in by this linter
}

/**
*	NewDefaultLinter returns a linter with absolutely no rules.
*
**/
func NewDefaultLinter() *Linter {
	return &Linter{}
}

/**
* Lint opens and lints the files and produces results that
* can be logged later on
**/
func (l *Linter) Lint(filepaths ...string) ([]*Result, []error) {
	var errors []error
	resources, errs := Read(filepaths...)
	for _, resource := range resources {
		l.resources = append(l.resources, &resource.Resource)
	}
	errors = append(errors, errs...)
	var results []*Result
	for _, resource := range resources {
		r, err := l.LintResource(resource)
		results = append(results, r...)
		errors = append(errors, err)
	}
	return results, errors
}

func (l *Linter) LintBytes(data []byte, filepath string) ([]*Result, []error) {
	resources, errors := ReadBytes(data, filepath)
	for _, resource := range resources {
		l.resources = append(l.resources, &resource.Resource)
	}
	var results []*Result
	for _, resource := range resources {
		r, err := l.LintResource(resource)
		results = append(results, r...)
		errors = append(errors, err)
	}
	return results, errors
}

func (l *Linter) LintFile(file *os.File) ([]*Result, []error) {
	resources, errors := ReadFile(file)
	for _, resource := range resources {
		l.resources = append(l.resources, &resource.Resource)
	}
	var results []*Result
	for _, resource := range resources {
		r, err := l.LintResource(resource)
		results = append(results, r...)
		errors = append(errors, err)
	}
	return results, errors
}

func (l *Linter) LintResource(resource *YamlDerivedResource) ([]*Result, error) {
	var results []*Result
	rules, err := l.createRules(&resource.Resource)
	ruleSorter := NewRuleSorter(rules)
	fixSorter := ruleSorter.Clone()
	l.fixes = append(l.fixes, fixSorter)
	for !ruleSorter.IsEmpty() {
		rule := ruleSorter.PopNextAvailable()
		if !rule.Condition() {
			results = append(results, &Result{
				Resource: resource,
				Message:  rule.Message,
				Level:    rule.Level,
			})
			dependentRules := ruleSorter.PopDependentRules(rule.ID)
			for _, dependentRule := range dependentRules {
				results = append(results, &Result{
					Resource: resource,
					Message:  dependentRule.Message,
					Level:    dependentRule.Level,
				})
			}
		} else {
			// this doesn't need to be fixed, so remove it from the fixSorter
			fixSorter.Remove(rule.ID)
		}
	}
	return results, err
}

/**
*	ApplyFixes applies all fixes that were registered as necessary during the lint phase.
*	The references to all the objects are kept in the Resources array so it will be reflected there.
**/
func (l *Linter) ApplyFixes() ([]*Resource, []string) {
	var appliedFixDescriptions []string
	for _, sorter := range l.fixes {
		for !sorter.IsEmpty() {
			rule := sorter.PopNextAvailable()
			fixed := rule.Fix()
			if !fixed {
				_ = sorter.PopDependentRules(rule.ID)
			} else {
				appliedFixDescriptions = append(appliedFixDescriptions, rule.FixDescription())
			}
		}
	}
	return l.resources, appliedFixDescriptions
}

/**
* createRules finds the type-appropriate rules that are registered in the linter
* and transforms them to generic rulese by applying the resource parameter.
* Then the list of rules are returned. I think I put it into a ruleSorter later on.
**/
func (l *Linter) createRules(resource *Resource) ([]*Rule, error) {
	var rules []*Rule

	switch concrete := resource.Object.(type) {
	case *appsv1.Deployment:
		for _, deploymentRule := range l.appsV1DeploymentRules {
			rules = append(rules, deploymentRule.CreateRule(concrete))
		}
	case *v1.Namespace:
		for _, v1NamespaceRule := range l.v1NamespaceRules {
			rules = append(rules, v1NamespaceRule.CreateRule(concrete))
		}
	case *v1.PersistentVolumeClaim:
		for _, v1PersistentVolumeClaimRule := range l.v1PersistentVolumeClaimRules {
			rules = append(rules, v1PersistentVolumeClaimRule.CreateRule(concrete))
		}
	case *v1beta1Extensions.Deployment:
		for _, v1Beta1ExtensionsDeploymentRule := range l.v1Beta1ExtensionsDeploymentRules {
			rules = append(rules, v1Beta1ExtensionsDeploymentRule.CreateRule(concrete))
		}
	case *batchV1.Job:
		for _, batchV1JobRule := range l.batchV1JobRules {
			rules = append(rules, batchV1JobRule.CreateRule(concrete))
		}
	case *batchV1beta1.CronJob:
		for _, batchV1Beta1CronJobRule := range l.batchV1Beta1CronJobRules {
			rules = append(rules, batchV1Beta1CronJobRule.CreateRule(concrete))
		}
	case *v1beta1Extensions.Ingress:
		for _, v1Beta1ExtensionsIngressRule := range l.v1Beta1ExtensionsIngressRules {
			rules = append(rules, v1Beta1ExtensionsIngressRule.CreateRule(concrete))
		}
	case *networkingV1.NetworkPolicy:
		for _, networkingV1NetworkPolicyRule := range l.networkingV1NetworkPolicyRules {
			rules = append(rules, networkingV1NetworkPolicyRule.CreateRule(concrete))
		}
	case *v1beta1Extensions.NetworkPolicy:
		for _, v1Beta1ExtensionsNetworkPolicyRule := range l.v1Beta1ExtensionsNetworkPolicyRules {
			rules = append(rules, v1Beta1ExtensionsNetworkPolicyRule.CreateRule(concrete))
		}
	case *rbacV1.Role:
		for _, rbacV1RoleRule := range l.rbacV1RoleRules {
			rules = append(rules, rbacV1RoleRule.CreateRule(concrete))
		}
	case *rbacV1beta1.RoleBinding:
		for _, rbacV1Beta1RoleBindingRule := range l.rbacV1Beta1RoleBindingRules {
			rules = append(rules, rbacV1Beta1RoleBindingRule.CreateRule(concrete))
		}
	case *v1.ServiceAccount:
		for _, v1ServiceAccountRule := range l.v1ServiceAccountRules {
			rules = append(rules, v1ServiceAccountRule.CreateRule(concrete))
		}
	case *v1.Service:
		for _, v1ServiceRule := range l.v1ServiceRules {
			rules = append(rules, v1ServiceRule.CreateRule(concrete))
		}

	default:
		return nil, fmt.Errorf("Resources of type %T have not been considered by the linter", concrete)
	}
	return rules, nil
}

/**
*	AddAppsV1DeploymentRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddAppsV1DeploymentRule(rules ...*AppsV1DeploymentRule) {
	l.appsV1DeploymentRules = append(l.appsV1DeploymentRules, rules...)
}

/**
*	AddV1NamespaceRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1NamespaceRule(rules ...*V1NamespaceRule) {
	l.v1NamespaceRules = append(l.v1NamespaceRules, rules...)
}

/**
*	AddV1PersistentVolumeClaimRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1PersistentVolumeClaimRule(rules ...*V1PersistentVolumeClaimRule) {
	l.v1PersistentVolumeClaimRules = append(l.v1PersistentVolumeClaimRules, rules...)
}

/**
*	AddV1Beta1ExtensionsDeploymentRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1Beta1ExtensionsDeploymentRule(rules ...*V1Beta1ExtensionsDeploymentRule) {
	l.v1Beta1ExtensionsDeploymentRules = append(l.v1Beta1ExtensionsDeploymentRules, rules...)
}

/**
*	AddBatchV1JobRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddBatchV1JobRule(rules ...*BatchV1JobRule) {
	l.batchV1JobRules = append(l.batchV1JobRules, rules...)
}

/**
*	AddBatchV1Beta1CronJobRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddBatchV1Beta1CronJobRule(rules ...*BatchV1Beta1CronJobRule) {
	l.batchV1Beta1CronJobRules = append(l.batchV1Beta1CronJobRules, rules...)
}

/**
*	AddV1Beta1ExtensionsIngressRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1Beta1ExtensionsIngressRule(rules ...*V1Beta1ExtensionsIngressRule) {
	l.v1Beta1ExtensionsIngressRules = append(l.v1Beta1ExtensionsIngressRules, rules...)
}

/**
*	AddNetworkingV1NetworkPolicyRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddNetworkingV1NetworkPolicyRule(rules ...*NetworkingV1NetworkPolicyRule) {
	l.networkingV1NetworkPolicyRules = append(l.networkingV1NetworkPolicyRules, rules...)
}

/**
*	AddV1Beta1ExtensionsNetworkPolicyRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1Beta1ExtensionsNetworkPolicyRule(rules ...*V1Beta1ExtensionsNetworkPolicyRule) {
	l.v1Beta1ExtensionsNetworkPolicyRules = append(l.v1Beta1ExtensionsNetworkPolicyRules, rules...)
}

/**
*	AddRbacV1RoleRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddRbacV1RoleRule(rules ...*RbacV1RoleRule) {
	l.rbacV1RoleRules = append(l.rbacV1RoleRules, rules...)
}

/**
*	AddRbacV1Beta1RoleBindingRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddRbacV1Beta1RoleBindingRule(rules ...*RbacV1Beta1RoleBindingRule) {
	l.rbacV1Beta1RoleBindingRules = append(l.rbacV1Beta1RoleBindingRules, rules...)
}

/**
*	AddV1ServiceAccountRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1ServiceAccountRule(rules ...*V1ServiceAccountRule) {
	l.v1ServiceAccountRules = append(l.v1ServiceAccountRules, rules...)
}

/**
*	AddV1ServiceRule adds a custom rule (or many) so that anything sent through the linter of the correct type
*	has this rule applied to it.
**/
func (l *Linter) AddV1ServiceRule(rules ...*V1ServiceRule) {
	l.v1ServiceRules = append(l.v1ServiceRules, rules...)
}
