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

type rule struct {
	ID             RuleID   // a string that uniquely identifies this rule wrt an object
	Prereqs        []RuleID // rules that this rule relies on for safe execution
	Condition      func() bool
	Message        string
	Level          log.Level // set the log level, only use this if you want to use logrus to help with logging.
	Resources      []*YamlDerivedResource
	Fix            func() bool // should mutate the underlying resource references in `Resources` somehow
	FixDescription func() string
}

// AppsV1DeploymentRule represents a semantic enforcement. For example, you would like all appsv1.Deployments to
// have 2 replicas. Your condition should check the field deployment.Spec.Replicas is non-nil and its value is 2.
// This represents a generic rule that can be applied to a deployment object.
// All other AppsV1DeploymentRule structs are analogous.
type AppsV1DeploymentRule struct {
	ID             RuleID                          // an arbitrary unique string identifier for this rule
	Prereqs        []RuleID                        // rules that should be executed before this rule (optional)
	Condition      func(*appsv1.Deployment) bool   // The Condition to execute on the deployment object. If this function returns true, it means that the deployment resource satisfies this rule.
	Message        string                          // The Message that should be reported to the user if the condition fails
	Level          log.Level                       // The level of severity implied if this rule fails
	Fix            func(*appsv1.Deployment) bool   // A mutating function that applies a fix. If Condition was called after this function was called, Condition should return true.
	FixDescription func(*appsv1.Deployment) string // A function returning the string that describes the fix that was applied within the Fix function
}

//	Once we get a reference to an actual resource, we can interpolate this into the
//	method bodies, and let every rule conform to the same structure.
//	At this point, we have no information about where this resource came from.
func (d *AppsV1DeploymentRule) createRule(deployment *appsv1.Deployment, ydr *YamlDerivedResource) *rule {
	r := &rule{
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

//	V1NamespaceRule represents a generic linter rule that can be applied to any v1.Namespace object.
type V1NamespaceRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.Namespace) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.Namespace) bool
	FixDescription func(*v1.Namespace) string
}

// createRule transforms a V1NamespaceRule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1NamespaceRule) createRule(namespace *v1.Namespace, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1PodSpecRule represents a generic linter rule that can be applied to any v1.Namespace object.
type V1PodSpecRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.PodSpec) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.PodSpec) bool
	FixDescription func(*v1.PodSpec) string
}

// createRule transforms a V1PodSpecRule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1PodSpecRule) createRule(podSpec *v1.PodSpec, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1ContainerRule represents a generic linter rule that can be applied to any v1.Namespace object.
type V1ContainerRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.Container) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.Container) bool
	FixDescription func(*v1.Container) string
}

// createRule transforms a V1ContainerRule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1ContainerRule) createRule(container *v1.Container, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1PersistentVolumeClaimRule represents a generic linter rule that can be applied to any v1.PersistentVolumeClaim object.
type V1PersistentVolumeClaimRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.PersistentVolumeClaim) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.PersistentVolumeClaim) bool
	FixDescription func(*v1.PersistentVolumeClaim) string
}

// createRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1PersistentVolumeClaimRule) createRule(pvc *v1.PersistentVolumeClaim, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1Beta1ExtensionsDeployment represents a generic linter rule that can be applied to any v1beta1Extensions.Deployment object.
type V1Beta1ExtensionsDeploymentRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1beta1Extensions.Deployment) bool
	Message        string
	Level          log.Level
	Fix            func(*v1beta1Extensions.Deployment) bool
	FixDescription func(*v1beta1Extensions.Deployment) string
}

// createRule transforms a V1Beta1ExtensionsDeploymentRule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1Beta1ExtensionsDeploymentRule) createRule(deployment *v1beta1Extensions.Deployment, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	BatchV1JobRule represents a generic linter rule that can be applied to any batchV1.Job object.
type BatchV1JobRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*batchV1.Job) bool
	Message        string
	Level          log.Level
	Fix            func(*batchV1.Job) bool
	FixDescription func(*batchV1.Job) string
}

// createRule transforms a BatchV1JobRule into a generic rule once it receives the parameter
// to interpolate.
func (r *BatchV1JobRule) createRule(job *batchV1.Job, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	BatchV1Beta1CronJobRule represents a generic linter rule that can be applied to any batchV1beta1.CronJob object.
type BatchV1Beta1CronJobRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*batchV1beta1.CronJob) bool
	Message        string
	Level          log.Level
	Fix            func(*batchV1beta1.CronJob) bool
	FixDescription func(*batchV1beta1.CronJob) string
}

// createRule transforms a BatchV1Beta1CronJobRule into a generic rule once it receives the parameter
// to interpolate.
func (r *BatchV1Beta1CronJobRule) createRule(cronjob *batchV1beta1.CronJob, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1Beta1ExtensionsIngressRule represents a generic linter rule that can be applied to any v1beta1Extensions.Ingress object.
type V1Beta1ExtensionsIngressRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1beta1Extensions.Ingress) bool
	Message        string
	Level          log.Level
	Fix            func(*v1beta1Extensions.Ingress) bool
	FixDescription func(*v1beta1Extensions.Ingress) string
}

// createRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1Beta1ExtensionsIngressRule) createRule(ingress *v1beta1Extensions.Ingress, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	NetworkingV1NetworkPolicyRule represents a generic linter rule that can be applied to any networkingV1.NetworkPolicy object.
type NetworkingV1NetworkPolicyRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*networkingV1.NetworkPolicy) bool
	Message        string
	Level          log.Level
	Fix            func(*networkingV1.NetworkPolicy) bool
	FixDescription func(*networkingV1.NetworkPolicy) string
}

// createRule transforms a NetworkingV1NetworkPolicyRule into a generic rule once it receives the parameter
// to interpolate.
func (r *NetworkingV1NetworkPolicyRule) createRule(networkpolicy *networkingV1.NetworkPolicy, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1Beta1ExtensionsNetworkPolicyRule represents a generic linter rule that can be applied to any v1beta1Extensions.NetworkPolicy object.
type V1Beta1ExtensionsNetworkPolicyRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1beta1Extensions.NetworkPolicy) bool
	Message        string
	Level          log.Level
	Fix            func(*v1beta1Extensions.NetworkPolicy) bool
	FixDescription func(*v1beta1Extensions.NetworkPolicy) string
}

// createRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1Beta1ExtensionsNetworkPolicyRule) createRule(networkpolicy *v1beta1Extensions.NetworkPolicy, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	RbacV1RoleRule represents a generic linter rule that can be applied to any rbacV1.Role object.
type RbacV1RoleRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*rbacV1.Role) bool
	Message        string
	Level          log.Level
	Fix            func(*rbacV1.Role) bool
	FixDescription func(*rbacV1.Role) string
}

// createRule transforms a RbacV1RoleRule into a generic rule once it receives the parameter
// to interpolate.
func (r *RbacV1RoleRule) createRule(role *rbacV1.Role, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	RbacV1Beta1RoleBindingRule represents a generic linter rule that can be applied to any rbacV1beta1.RoleBinding object.
type RbacV1Beta1RoleBindingRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*rbacV1beta1.RoleBinding) bool
	Message        string
	Level          log.Level
	Fix            func(*rbacV1beta1.RoleBinding) bool
	FixDescription func(*rbacV1beta1.RoleBinding) string
}

// createRule transforms a RbacV1Beta1RoleBindingRule into a generic rule once it receives the parameter
// to interpolate.
func (r *RbacV1Beta1RoleBindingRule) createRule(rolebinding *rbacV1beta1.RoleBinding, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1ServiceAccountRule represents a generic linter rule that can be applied to any v1.ServiceAccount object.
type V1ServiceAccountRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.ServiceAccount) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.ServiceAccount) bool
	FixDescription func(*v1.ServiceAccount) string
}

// createRule transforms a <ResourceType>Rule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1ServiceAccountRule) createRule(serviceaccount *v1.ServiceAccount, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	V1ServiceRule represents a generic linter rule that can be applied to any v1.Service object.
type V1ServiceRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*v1.Service) bool
	Message        string
	Level          log.Level
	Fix            func(*v1.Service) bool
	FixDescription func(*v1.Service) string
}

// createRule transforms a V1ServiceRule into a generic rule once it receives the parameter
// to interpolate.
func (r *V1ServiceRule) createRule(service *v1.Service, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	GenericRule represents a generic linter rule that can be applied to an object of any type.
//	Use this if the type you want to apply a check to is not currently supported, or it's a check
//	that can apply uniformly to all resources, for example, each resource is registered under a namespace.
type GenericRule struct {
	ID             RuleID
	Prereqs        []RuleID
	Condition      func(*Resource) bool
	Message        string
	Level          log.Level
	Fix            func(*Resource) bool
	FixDescription func(*Resource) string
}

// createRule transforms a GenericRule into a generic rule once it receives the parameter
// to interpolate.
func (r *GenericRule) createRule(resource *Resource, ydr *YamlDerivedResource) *rule {
	rule := &rule{
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

//	InterdependentRule represents a generic linter rule that will be applied to the resources as a whole.
//	An example would be to check that for all objects, their namespace corresponds to an existing namespace object.
//	You will need to do your own typecasting or rely on the methods available to you in metav1.Object and meta.Type to access the objects' fields.
type InterdependentRule struct {
	ID             RuleID
	Condition      func([]*Resource) (bool, []*Resource) // if it returns false, it will also return a list of the offending resources. This is passed to the result.Resources field later.
	Message        string
	Level          log.Level
	Fix            func([]*Resource) bool
	FixDescription func([]*Resource) string
}

type interdependentRule struct {
	ID             RuleID
	Condition      func() bool // if it returns false, it will also return a list of the offending resources. This is passed to the result.Resources field later.
	Message        string
	Level          log.Level
	Fix            func() bool
	FixDescription func() string
	Resources      []*YamlDerivedResource
}

// createRule transforms a InterdependentRule into a generic rule once it receives the parameter
// to interpolate.
func (r *InterdependentRule) createRule(resources []*YamlDerivedResource) *interdependentRule {
	var bareResources []*Resource
	for _, r := range resources {
		bareResources = append(bareResources, &r.Resource)
	}
	// we need to silently execute the condition so we can find out which resources are relevant :(
	// This means prerequisites are disallowed. sorry :(
	success, offendingResources := r.Condition(bareResources)
	// collect information about offending resources
	var offendingYamls []*YamlDerivedResource
	for _, offendingResource := range offendingResources {
		for _, yaml := range resources {
			if &yaml.Resource == offendingResource {
				offendingYamls = append(offendingYamls, yaml)
			}
		}
	}
	rule := &interdependentRule{
		ID: r.ID,
		Condition: func() bool {
			return success
		},
		Message:   r.Message,
		Level:     r.Level,
		Resources: offendingYamls,
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
