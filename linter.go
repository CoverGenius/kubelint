package kubelint

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

// Linter: This Linter represents something you can pass in resources to
//	and get results out of that you can eventually log.
//	Also some utility methods for input handling.
type Linter struct {
	logger                              *log.Logger
	appsV1DeploymentRules               []*AppsV1DeploymentRule               // a register for all user-defined appsV1Deployment rules
	v1NamespaceRules                    []*V1NamespaceRule                    // a register for all user-defined v1Namespace rules
	v1PodSpecRules                      []*V1PodSpecRule                      // a register for all user-defined v1PodSpec rules
	v1ContainerRules                    []*V1ContainerRule                    // a register for all user-defined v1Container rules
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
	genericRules                        []*GenericRule                        // a register for all user-defined Generic rules (applied to every object)
	interdependentRules                 []*InterdependentRule                 // a register for all user-defined Interdependent rules (applied to the system as a whole)
	fixes                               []*ruleSorter                         // fixes that should be applied to the resources in order to mitigate some errors on a future pass
	resources                           []*Resource                           // All the resources that have been read in by this linter
}

//	NewDefaultLinter returns a linter with absolutely no rules.
func NewDefaultLinter() *Linter {
	return &Linter{}
}

func NewLinter(l *log.Logger) *Linter {
	return &Linter{logger: l}
}

// Lint opens and lints the files and produces results that
// can be logged later on
func (l *Linter) Lint(filepaths ...string) ([]*Result, []error) {
	l.logger.Debugf("Linting files: %#v\n", filepaths)
	var errors []error
	resources, errs := Read(filepaths...)
	for _, resource := range resources {
		l.resources = append(l.resources, &resource.Resource)
	}
	errors = append(errors, errs...)
	var results []*Result
	// add interdependent checks
	results = append(results, l.lintResources(resources)...)
	for _, resource := range resources {
		r, err := l.LintResource(resource)
		l.logger.Debugln("results from linting", resource.Filepath, r)
		results = append(results, r...)
		if err != nil {
			l.logger.Debugln("Error from LintResource: ", err)
			errors = append(errors, err)
		}
	}
	return results, errors
}

//	LintBytes takes a slice of bytes to lint and a filepath and
//	returns a list of Results and errors to report or log later on
func (l *Linter) LintBytes(data []byte, filepath string) ([]*Result, []error) {
	resources, errors := ReadBytes(data, filepath)
	for _, resource := range resources {
		l.resources = append(l.resources, &resource.Resource)
	}
	var results []*Result
	// add interdependent checks
	results = append(results, l.lintResources(resources)...)
	for _, resource := range resources {
		r, err := l.LintResource(resource)
		results = append(results, r...)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return results, errors
}

//	LintFile takes a file pointer and returns a list of Reults and Errors
//	to be logged or reported later on
func (l *Linter) LintFile(file *os.File) ([]*Result, []error) {
	resources, errors := ReadFile(file)
	for _, resource := range resources {
		l.resources = append(l.resources, &resource.Resource)
	}
	var results []*Result
	// add interdependent checks
	results = append(results, l.lintResources(resources)...)
	for _, resource := range resources {
		r, err := l.LintResource(resource)
		results = append(results, r...)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return results, errors
}

//	lintResources takes a list of Yaml Derived Resources, applying interdependent rules ONLY
//   and returns a list of Results
//	to be logged or reported
func (l *Linter) lintResources(resources []*YamlDerivedResource) []*Result {
	var results []*Result
	rules := l.createInterdependentRules(resources)
	ruleSorter := newRuleSorter(rules)
	fixSorter := ruleSorter.clone()
	l.fixes = append(l.fixes, fixSorter)
	for !ruleSorter.isEmpty() {
		rule := ruleSorter.popNextAvailable()
		if !rule.Condition() {
			results = append(results, &Result{
				Resources: resources,
				Message:   rule.Message,
				Level:     rule.Level,
			})
			dependentRules := ruleSorter.popDependentRules(rule.ID)
			for _, dependentRule := range dependentRules {
				results = append(results, &Result{
					Resources: resources,
					Message:   dependentRule.Message,
					Level:     dependentRule.Level,
				})
			}
		} else {
			fixSorter.remove(rule.ID)
		}
	}
	return results
}

// LintResource takes a yaml derived resource and returns a list of results and errors
// to be logged or reported
func (l *Linter) LintResource(resource *YamlDerivedResource) ([]*Result, error) {
	var results []*Result
	rules, err := l.createRules(resource)
	l.logger.Debugln(len(rules), "rules created for", resource.Filepath)
	// log rules and their dependent rules
	for _, rule := range rules {
		l.logger.Debugf("Rule ID: %s\n\tPrereqs: %#v\n", rule.ID, rule.Prereqs)
	}
	ruleSorter := newRuleSorter(rules)
	fixSorter := ruleSorter.clone()
	l.fixes = append(l.fixes, fixSorter)
	for !ruleSorter.isEmpty() {
		rule := ruleSorter.popNextAvailable()
		l.logger.Debugln("Testing rule", rule.ID)
		if !rule.Condition() {
			results = append(results, &Result{
				Resources: []*YamlDerivedResource{resource},
				Message:   rule.Message,
				Level:     rule.Level,
			})
			dependentRules := ruleSorter.popDependentRules(rule.ID)
			for _, dependentRule := range dependentRules {
				results = append(results, &Result{
					Resources: []*YamlDerivedResource{resource},
					Message:   dependentRule.Message,
					Level:     dependentRule.Level,
				})
			}
		} else {
			// this doesn't need to be fixed, so remove it from the fixSorter
			fixSorter.remove(rule.ID)
		}
	}
	return results, err
}

//	ApplyFixes applies all fixes that were registered as necessary during the lint phase.
//	The references to all the objects are kept in the Resources array so it will be reflected there.
func (l *Linter) ApplyFixes() ([]*Resource, []string) {
	var appliedFixDescriptions []string
	for _, sorter := range l.fixes {
		for !sorter.isEmpty() {
			rule := sorter.popNextAvailable()
			fixed := rule.Fix()
			if !fixed {
				_ = sorter.popDependentRules(rule.ID)
			} else {
				appliedFixDescriptions = append(appliedFixDescriptions, rule.FixDescription())
			}
		}
	}
	return l.resources, appliedFixDescriptions
}

//	CreateRules finds the registered interdependent rules and transforms them
//	to generic rules by applying the ydrs parameter.
func (l *Linter) createInterdependentRules(ydrs []*YamlDerivedResource) []*Rule {
	var rules []*Rule
	for _, interdependentRule := range l.interdependentRules {
		rules = append(rules, interdependentRule.CreateRule(ydrs))
	}
	return rules
}

// createRules finds the type-appropriate rules that are registered in the linter
// and transforms them to generic rules by applying the resource parameter.
// Then the list of rules are returned. I think I put it into a ruleSorter later on.
func (l *Linter) createRules(ydr *YamlDerivedResource) ([]*Rule, error) {
	var rules []*Rule
	resource := &ydr.Resource

	// generic rules always need to be added
	for _, genericRule := range l.genericRules {
		rules = append(rules, genericRule.CreateRule(resource, ydr))
	}
	// append type-specific rules
	switch concrete := resource.Object.(type) {
	case *appsv1.Deployment:
		for _, deploymentRule := range l.appsV1DeploymentRules {
			rules = append(rules, deploymentRule.CreateRule(concrete, ydr))
		}
		for _, podSpecRule := range l.v1PodSpecRules {
			rules = append(rules, podSpecRule.CreateRule(&concrete.Spec.Template.Spec, ydr))
		}
		for _, v1ContainerRule := range l.v1ContainerRules {
			for i, _ := range concrete.Spec.Template.Spec.Containers {
				rules = append(rules, v1ContainerRule.CreateRule(&concrete.Spec.Template.Spec.Containers[i], ydr))
			}
		}
	case *v1.Namespace:
		for _, v1NamespaceRule := range l.v1NamespaceRules {
			rules = append(rules, v1NamespaceRule.CreateRule(concrete, ydr))
		}
	case *v1.PersistentVolumeClaim:
		for _, v1PersistentVolumeClaimRule := range l.v1PersistentVolumeClaimRules {
			rules = append(rules, v1PersistentVolumeClaimRule.CreateRule(concrete, ydr))
		}
	case *v1beta1Extensions.Deployment:
		for _, v1Beta1ExtensionsDeploymentRule := range l.v1Beta1ExtensionsDeploymentRules {
			rules = append(rules, v1Beta1ExtensionsDeploymentRule.CreateRule(concrete, ydr))
		}
	case *batchV1.Job:
		for _, batchV1JobRule := range l.batchV1JobRules {
			rules = append(rules, batchV1JobRule.CreateRule(concrete, ydr))
		}
	case *batchV1beta1.CronJob:
		for _, batchV1Beta1CronJobRule := range l.batchV1Beta1CronJobRules {
			rules = append(rules, batchV1Beta1CronJobRule.CreateRule(concrete, ydr))
		}
	case *v1beta1Extensions.Ingress:
		for _, v1Beta1ExtensionsIngressRule := range l.v1Beta1ExtensionsIngressRules {
			rules = append(rules, v1Beta1ExtensionsIngressRule.CreateRule(concrete, ydr))
		}
	case *networkingV1.NetworkPolicy:
		for _, networkingV1NetworkPolicyRule := range l.networkingV1NetworkPolicyRules {
			rules = append(rules, networkingV1NetworkPolicyRule.CreateRule(concrete, ydr))
		}
	case *v1beta1Extensions.NetworkPolicy:
		for _, v1Beta1ExtensionsNetworkPolicyRule := range l.v1Beta1ExtensionsNetworkPolicyRules {
			rules = append(rules, v1Beta1ExtensionsNetworkPolicyRule.CreateRule(concrete, ydr))
		}
	case *rbacV1.Role:
		for _, rbacV1RoleRule := range l.rbacV1RoleRules {
			rules = append(rules, rbacV1RoleRule.CreateRule(concrete, ydr))
		}
	case *rbacV1beta1.RoleBinding:
		for _, rbacV1Beta1RoleBindingRule := range l.rbacV1Beta1RoleBindingRules {
			rules = append(rules, rbacV1Beta1RoleBindingRule.CreateRule(concrete, ydr))
		}
	case *v1.ServiceAccount:
		for _, v1ServiceAccountRule := range l.v1ServiceAccountRules {
			rules = append(rules, v1ServiceAccountRule.CreateRule(concrete, ydr))
		}
	case *v1.Service:
		for _, v1ServiceRule := range l.v1ServiceRules {
			rules = append(rules, v1ServiceRule.CreateRule(concrete, ydr))
		}

	default:
		return nil, fmt.Errorf("Resources of type %T have not been considered by the linter", concrete)
	}
	return rules, nil
}

//	AddAppsV1DeploymentRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddAppsV1DeploymentRule(rules ...*AppsV1DeploymentRule) {
	l.appsV1DeploymentRules = append(l.appsV1DeploymentRules, rules...)
}

//	AddV1NamespaceRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1NamespaceRule(rules ...*V1NamespaceRule) {
	l.v1NamespaceRules = append(l.v1NamespaceRules, rules...)
}

//	AddV1PodSpecRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1PodSpecRule(rules ...*V1PodSpecRule) {
	l.v1PodSpecRules = append(l.v1PodSpecRules, rules...)
}

//	AddV1ContainerRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1ContainerRule(rules ...*V1ContainerRule) {
	l.v1ContainerRules = append(l.v1ContainerRules, rules...)
}

//	AddV1PersistentVolumeClaimRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1PersistentVolumeClaimRule(rules ...*V1PersistentVolumeClaimRule) {
	l.v1PersistentVolumeClaimRules = append(l.v1PersistentVolumeClaimRules, rules...)
}

//	AddV1Beta1ExtensionsDeploymentRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1Beta1ExtensionsDeploymentRule(rules ...*V1Beta1ExtensionsDeploymentRule) {
	l.v1Beta1ExtensionsDeploymentRules = append(l.v1Beta1ExtensionsDeploymentRules, rules...)
}

//	AddBatchV1JobRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddBatchV1JobRule(rules ...*BatchV1JobRule) {
	l.batchV1JobRules = append(l.batchV1JobRules, rules...)
}

//	AddBatchV1Beta1CronJobRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddBatchV1Beta1CronJobRule(rules ...*BatchV1Beta1CronJobRule) {
	l.batchV1Beta1CronJobRules = append(l.batchV1Beta1CronJobRules, rules...)
}

//	AddV1Beta1ExtensionsIngressRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1Beta1ExtensionsIngressRule(rules ...*V1Beta1ExtensionsIngressRule) {
	l.v1Beta1ExtensionsIngressRules = append(l.v1Beta1ExtensionsIngressRules, rules...)
}

//	AddNetworkingV1NetworkPolicyRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddNetworkingV1NetworkPolicyRule(rules ...*NetworkingV1NetworkPolicyRule) {
	l.networkingV1NetworkPolicyRules = append(l.networkingV1NetworkPolicyRules, rules...)
}

//	AddV1Beta1ExtensionsNetworkPolicyRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1Beta1ExtensionsNetworkPolicyRule(rules ...*V1Beta1ExtensionsNetworkPolicyRule) {
	l.v1Beta1ExtensionsNetworkPolicyRules = append(l.v1Beta1ExtensionsNetworkPolicyRules, rules...)
}

//	AddRbacV1RoleRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddRbacV1RoleRule(rules ...*RbacV1RoleRule) {
	l.rbacV1RoleRules = append(l.rbacV1RoleRules, rules...)
}

//	AddRbacV1Beta1RoleBindingRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddRbacV1Beta1RoleBindingRule(rules ...*RbacV1Beta1RoleBindingRule) {
	l.rbacV1Beta1RoleBindingRules = append(l.rbacV1Beta1RoleBindingRules, rules...)
}

//	AddV1ServiceAccountRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1ServiceAccountRule(rules ...*V1ServiceAccountRule) {
	l.v1ServiceAccountRules = append(l.v1ServiceAccountRules, rules...)
}

//	AddV1ServiceRule adds a custom rule (or many) so that anything sent through the linter of the correct type
//	has this rule applied to it.
func (l *Linter) AddV1ServiceRule(rules ...*V1ServiceRule) {
	l.v1ServiceRules = append(l.v1ServiceRules, rules...)
}

//	AddGenericRule adds a custom rule (or many) so that anything sent through the linter
//	has this rule applied to it.
func (l *Linter) AddGenericRule(rules ...*GenericRule) {
	l.genericRules = append(l.genericRules, rules...)
}

//	AddInterdependentRule adds a custom rule (or many) so that anything sent through the linter
//	has this rule applied to it.
func (l *Linter) AddInterdependentRule(rules ...*InterdependentRule) {
	l.interdependentRules = append(l.interdependentRules, rules...)
}
