package kubelint

import (
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	batchV1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	v1beta1Extensions "k8s.io/api/extensions/v1beta1"
	networkingV1 "k8s.io/api/networking/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	rbacV1beta1 "k8s.io/api/rbac/v1beta1"
)

// The unique identifier for a rule. This lets us define an execution order with the Prereqs field.
type RuleID string

type Rule struct {
	ID             RuleID   // a string that uniquely identifies this rule wrt an object
	Prereqs        []RuleID // rules that this rule relies on for safe execution
	Condition      func() bool
	Message        string
	Level          log.Level // set the log level, only use this if you want to use logrus to help with logging.
	Resources      []*YamlDerivedResource
	Fix            func() bool // should mutate the underlying resource references in `Resources` somehow
	FixDescription func() string
}

/**
*	This represents a generic rule that can be applied to a deployment object.
* 	All other AppsV1DeploymentRule structs are analogous.
**/
type AppsV1DeploymentRule struct {
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
func (d *AppsV1DeploymentRule) CreateRule(deployment *appsv1.Deployment, ydr *YamlDerivedResource) *Rule {
	r := &Rule{
		ID:      d.ID,
		Prereqs: d.Prereqs,
		Condition: func() bool {
			if d.Condition == nil {
				return true
			}
			return d.Condition(deployment)
		},
		Message:   d.Message,
		Level:     d.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if d.Fix == nil {
				return false
			}
			return d.Fix(deployment)
		},
		FixDescription: func() string {
			if d.FixDescription == nil {
				return ""
			}
			return d.FixDescription(deployment)
		},
	}
	return r
}

/**
*	V1NamespaceRule represents a generic linter rule that can be applied to any v1.Namespace object.
**/
type V1NamespaceRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.Namespace) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.Namespace) bool
	FixDescription func(*v1.Namespace) string
}

/**
* CreateRule transforms a V1NamespaceRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1NamespaceRule) CreateRule(namespace *v1.Namespace, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(namespace)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},

		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(namespace)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(namespace)
		},
	}
	return rule
}

/**
*	V1PodSpecRule represents a generic linter rule that can be applied to any v1.Namespace object.
**/
type V1PodSpecRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.PodSpec) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.PodSpec) bool
	FixDescription func(*v1.PodSpec) string
}

/**
* CreateRule transforms a V1PodSpecRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1PodSpecRule) CreateRule(podSpec *v1.PodSpec, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(podSpec)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},

		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(podSpec)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(podSpec)
		},
	}
	return rule
}

/**
*	V1ContainerRule represents a generic linter rule that can be applied to any v1.Namespace object.
**/
type V1ContainerRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.Container) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.Container) bool
	FixDescription func(*v1.Container) string
}

/**
* CreateRule transforms a V1ContainerRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1ContainerRule) CreateRule(container *v1.Container, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(container)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},

		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(container)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(container)
		},
	}
	return rule
}

/**
*	V1PersistentVolumeClaimRule represents a generic linter rule that can be applied to any v1.PersistentVolumeClaim object.
**/
type V1PersistentVolumeClaimRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.PersistentVolumeClaim) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.PersistentVolumeClaim) bool
	FixDescription func(*v1.PersistentVolumeClaim) string
}

/**
* CreateRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1PersistentVolumeClaimRule) CreateRule(pvc *v1.PersistentVolumeClaim, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(pvc)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(pvc)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(pvc)
		},
	}
	return rule
}

/**
*	V1Beta1ExtensionsDeployment represents a generic linter rule that can be applied to any v1beta1Extensions.Deployment object.
**/
type V1Beta1ExtensionsDeploymentRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1beta1Extensions.Deployment) bool
	Message        string
	Level          log.Level
	Fix            func(*v1beta1Extensions.Deployment) bool
	FixDescription func(*v1beta1Extensions.Deployment) string
}

/**
* CreateRule transforms a V1Beta1ExtensionsDeploymentRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1Beta1ExtensionsDeploymentRule) CreateRule(deployment *v1beta1Extensions.Deployment, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(deployment)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(deployment)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(deployment)
		},
	}
	return rule
}

/**
*	BatchV1JobRule represents a generic linter rule that can be applied to any batchV1.Job object.
**/
type BatchV1JobRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*batchV1.Job) bool
	Message        string
	Level          log.Level
	Fix            func(*batchV1.Job) bool
	FixDescription func(*batchV1.Job) string
}

/**
* CreateRule transforms a BatchV1JobRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *BatchV1JobRule) CreateRule(job *batchV1.Job, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(job)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(job)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(job)
		},
	}
	return rule
}

/**
*	BatchV1Beta1CronJobRule represents a generic linter rule that can be applied to any batchV1beta1.CronJob object.
**/
type BatchV1Beta1CronJobRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*batchV1beta1.CronJob) bool
	Message        string
	Level          log.Level
	Fix            func(*batchV1beta1.CronJob) bool
	FixDescription func(*batchV1beta1.CronJob) string
}

/**
* CreateRule transforms a BatchV1Beta1CronJobRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *BatchV1Beta1CronJobRule) CreateRule(cronjob *batchV1beta1.CronJob, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(cronjob)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(cronjob)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(cronjob)
		},
	}
	return rule
}

/**
*	V1Beta1ExtensionsIngressRule represents a generic linter rule that can be applied to any v1beta1Extensions.Ingress object.
**/
type V1Beta1ExtensionsIngressRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1beta1Extensions.Ingress) bool
	Message        string
	Level          log.Level
	Fix            func(*v1beta1Extensions.Ingress) bool
	FixDescription func(*v1beta1Extensions.Ingress) string
}

/**
* CreateRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1Beta1ExtensionsIngressRule) CreateRule(ingress *v1beta1Extensions.Ingress, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(ingress)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(ingress)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(ingress)
		},
	}
	return rule
}

/**
*	NetworkingV1NetworkPolicyRule represents a generic linter rule that can be applied to any networkingV1.NetworkPolicy object.
**/
type NetworkingV1NetworkPolicyRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*networkingV1.NetworkPolicy) bool
	Message        string
	Level          log.Level
	Fix            func(*networkingV1.NetworkPolicy) bool
	FixDescription func(*networkingV1.NetworkPolicy) string
}

/**
* CreateRule transforms a NetworkingV1NetworkPolicyRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *NetworkingV1NetworkPolicyRule) CreateRule(networkpolicy *networkingV1.NetworkPolicy, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(networkpolicy)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(networkpolicy)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(networkpolicy)
		},
	}
	return rule
}

/**
*	V1Beta1ExtensionsNetworkPolicyRule represents a generic linter rule that can be applied to any v1beta1Extensions.NetworkPolicy object.
**/
type V1Beta1ExtensionsNetworkPolicyRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1beta1Extensions.NetworkPolicy) bool
	Message        string
	Level          log.Level
	Fix            func(*v1beta1Extensions.NetworkPolicy) bool
	FixDescription func(*v1beta1Extensions.NetworkPolicy) string
}

/**
* CreateRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1Beta1ExtensionsNetworkPolicyRule) CreateRule(networkpolicy *v1beta1Extensions.NetworkPolicy, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(networkpolicy)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(networkpolicy)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(networkpolicy)
		},
	}
	return rule
}

/**
*	RbacV1RoleRule represents a generic linter rule that can be applied to any rbacV1.Role object.
**/
type RbacV1RoleRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*rbacV1.Role) bool
	Message        string
	Level          log.Level
	Fix            func(*rbacV1.Role) bool
	FixDescription func(*rbacV1.Role) string
}

/**
* CreateRule transforms a RbacV1RoleRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *RbacV1RoleRule) CreateRule(role *rbacV1.Role, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(role)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(role)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(role)
		},
	}
	return rule
}

/**
*	RbacV1Beta1RoleBindingRule represents a generic linter rule that can be applied to any rbacV1beta1.RoleBinding object.
**/
type RbacV1Beta1RoleBindingRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*rbacV1beta1.RoleBinding) bool
	Message        string
	Level          log.Level
	Fix            func(*rbacV1beta1.RoleBinding) bool
	FixDescription func(*rbacV1beta1.RoleBinding) string
}

/**
* CreateRule transforms a RbacV1Beta1RoleBindingRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *RbacV1Beta1RoleBindingRule) CreateRule(rolebinding *rbacV1beta1.RoleBinding, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(rolebinding)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(rolebinding)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(rolebinding)
		},
	}
	return rule
}

/**
*	V1ServiceAccountRule represents a generic linter rule that can be applied to any v1.ServiceAccount object.
**/
type V1ServiceAccountRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.ServiceAccount) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.ServiceAccount) bool
	FixDescription func(*v1.ServiceAccount) string
}

/**
* CreateRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1ServiceAccountRule) CreateRule(serviceaccount *v1.ServiceAccount, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(serviceaccount)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(serviceaccount)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(serviceaccount)
		},
	}
	return rule
}

/**
*	V1ServiceRule represents a generic linter rule that can be applied to any v1.Service object.
**/
type V1ServiceRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.Service) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.Service) bool
	FixDescription func(*v1.Service) string
}

/**
* CreateRule transforms a V1ServiceRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *V1ServiceRule) CreateRule(service *v1.Service, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(service)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(service)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(service)
		},
	}
	return rule
}

/**
*	GenericRule represents a generic linter rule that can be applied to an object of any type.
*	Use this if the type you want to apply a check to is not currently supported, or it's a check
*	that can apply uniformly to all resources, for example, each resource is registered under a namespace.
**/
type GenericRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*Resource) bool
	Message        string
	Level          log.Level
	Fix            func(*Resource) bool
	FixDescription func(*Resource) string
}

/**
* CreateRule transforms a GenericRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *GenericRule) CreateRule(resource *Resource, ydr *YamlDerivedResource) *Rule {
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(resource)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: []*YamlDerivedResource{ydr},
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(resource)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(resource)
		},
	}
	return rule
}

/**
*	InterdependentRule represents a generic linter rule that will be applied to the resources as a whole.
*	An example would be to check that for all objects, their namespace corresponds to an existing namespace object.
*	You will need to do your own typecasting or rely on the methods available to you in metav1.Object and meta.Type to access the objects' fields.
**/
type InterdependentRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func([]*Resource) bool
	Message        string
	Level          log.Level
	Fix            func([]*Resource) bool
	FixDescription func([]*Resource) string
}

/**
* CreateRule transforms a InterdependentRule into a generic rule once it receives the parameter
* to interpolate.
**/
func (r *InterdependentRule) CreateRule(resources []*YamlDerivedResource) *Rule {
	var bareResources []*Resource
	for _, r := range resources {
		bareResources = append(bareResources, &r.Resource)
	}
	rule := &Rule{
		ID:      r.ID,
		Prereqs: r.Prereqs,
		Condition: func() bool {
			if r.Condition == nil {
				return true
			}
			return r.Condition(bareResources)
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: resources,
		Fix: func() bool {
			if r.Fix == nil {
				return false
			}
			return r.Fix(bareResources)
		},
		FixDescription: func() string {
			if r.FixDescription == nil {
				return ""
			}
			return r.FixDescription(bareResources)
		},
	}
	return rule
}
