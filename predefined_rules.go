package kubelint

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	batchV1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
)

/*

Predefined rules relating to resources of type appsv1.Deployment

- An AppsV1Deployment should have a project label: APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL

- An AppsV1Deployment should have an app.kubernetes.io/name label: APPSV1_DEPLOYMENT_EXISTS_APP_K8S_LABEL

- An AppsV1Deployment should be within a namespace: APPSV1_DEPLOYMENT_WITHIN_NAMESPACE

- An AppsV1Deployment should specify a liveness endpoint: APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS

- An AppsV1Deployment should specify a readiness endpoint: APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS

- An AppsV1Deploument should have liveness and readiness endpoints that aren't the same: APPSV1_DEPLOYMENT_LIVENESS_READINESS_NONMATCHING

Predefined rules relating to resources of type v1.PodSpec

- A V1PodSpec should have a non-nil security context: V1_PODSPEC_NON_NIL_SECURITY_CONTEXT

- A V1PodSpec should specify runAsNonRoot: true: V1_PODSPEC_RUN_AS_NON_ROOT

- A V1PodSpec should have a user and group ID of 44444: V1_PODSPEC_CORRECT_USER_GROUP_ID

- A V1PodSpec should have exactly one container: V1_PODSPEC_EXACTLY_1_CONTAINER

- A V1PodSpec should have a non-zero number of containers: V1_PODSPEC_NON_ZERO_CONTAINERS

Predefined rules relating to resources of type v1.Container

- A V1Container should have a non-null security context: V1_CONTAINER_EXISTS_SECURITY_CONTEXT

- A V1Container should not allow privilege escalation: V1_CONTAINER_ALLOW_PRIVILEGE_ESCALATION_FALSE

- A V1Container's image should come from a set of allowed images defined in isImageAllowed: V1_CONTAINER_VALID_IMAGE

- A V1Container should have privileged set to false: V1_CONTAINER_PRIVILEGED_FALSE

- A V1Container should specify Resource Limits and Requests: V1_CONTAINER_EXISTS_RESOURCE_LIMITS_AND_REQUESTS

- A V1Container should make CPU requests that are less than or equal to 100%: V1_CONTAINER_REQUESTS_CPU_REASONABLE

Predefined rules relating to resources of type batchV1.Job

- A BatchV1Beta1CronJob should be within a namespace

- A BatchV1Beta1CronJob should forbid concurrent operations: BATCHV1_BETA1_CRONJOB_FORBID_CONCURRENT

- A BatchV1Job should be within a namespace: BATCHV1_JOB_WITHIN_NAMESPACE

- A BatchV1Job's restart policy should be set to Never: BATCHV1_JOB_RESTART_NEVER

- A BatchV1Job's Time to Live should be set: BATCHV1_JOB_EXISTS_TTL

Predefined rules relating to resources of type v1.Namespace

- A V1Namespace should have a valid DNS name: V1_NAMESPACE_VALID_DNS

Predefined rules relating to resources of type v1.Service

- A V1Service should be within a namespace: V1_SERVICE_WITHIN_NAMESPACE

- A V1Service name should be a valid DNS: V1_SERVICE_NAME_VALID_DNS

Predefined interdependent rules

- A unit should always contain one namespace: INTERDEPENDENT_ONE_NAMESPACE

- All resources should be under the namespace in the unit: INTERDEPENDENT_MATCHING_NAMESPACE

- The unit should contain a network policy: INTERDEPENDENT_NETWORK_POLICY_REQUIRED
*/
var (
	// An AppsV1Deployment should have a project label.
	APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL = &AppsV1DeploymentRule{
		ID: "APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL",
		Condition: func(deployment *appsv1.Deployment) bool {
			_, found := deployment.Spec.Template.Labels["project"]
			return found
		},
		Message: "There should be a project label present under the deployment's spec.template.labels",
		Level:   log.ErrorLevel,
	}
	// An AppsV1Deployment should have an app.kubernetes.io/name label.
	APPSV1_DEPLOYMENT_EXISTS_APP_K8S_LABEL = &AppsV1DeploymentRule{
		ID: "APPSV1_DEPLOYMENT_EXISTS_APP_K8S_LABEL",
		Condition: func(deployment *appsv1.Deployment) bool {
			_, found := deployment.Spec.Template.Labels["app.kubernetes.io/name"]
			return found
		},
		Message: "There should be an app.kubernetes.io/name label present for the deployment's spec.template",
		Level:   log.ErrorLevel,
		Fix: func(deployment *appsv1.Deployment) bool {
			label, found := deployment.Spec.Template.Labels["app"]
			if found {
				delete(deployment.Spec.Template.Labels, "app")
				deployment.Spec.Template.Labels["app.kubernetes.io/name"] = label
				return true
			}
			return false
		},
		FixDescription: func(deployment *appsv1.Deployment) string {
			return fmt.Sprintf("Found app label in deployment %s and used this value to populate the \"app.kubernetes.io/name\" key", deployment.Name)
		},
	}
	// An AppsV1Deployment should be within a namespace
	APPSV1_DEPLOYMENT_WITHIN_NAMESPACE = &AppsV1DeploymentRule{
		ID: "APPSV1_DEPLOYMENT_WITHIN_NAMESPACE",
		Condition: func(deployment *appsv1.Deployment) bool {
			return deployment.Namespace != ""
		},
		Message: "The resource must be within a namespace",
		Level:   log.ErrorLevel,
	}
	// An AppsV1Deployment should specify a liveness endpoint
	APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS = &AppsV1DeploymentRule{
		ID:      "APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS",
		Prereqs: []RuleID{"V1_PODSPEC_NON_ZERO_CONTAINERS"},
		Condition: func(deployment *appsv1.Deployment) bool {
			return deployment.Spec.Template.Spec.Containers[0].LivenessProbe != nil &&
				deployment.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.HTTPGet != nil
		},
		Message: "Expected declaration of liveness probe for the container (livenessProbe)",
		Level:   log.ErrorLevel,
	}
	// An AppsV1Deployment should specify a readiness endpoint
	APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS = &AppsV1DeploymentRule{
		ID:      "APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS",
		Prereqs: []RuleID{"V1_PODSPEC_NON_ZERO_CONTAINERS"},
		Condition: func(deployment *appsv1.Deployment) bool {
			return deployment.Spec.Template.Spec.Containers[0].ReadinessProbe != nil &&
				deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.HTTPGet != nil
		},
		Message: "Expected declaration of readiness probe for the container (readinessProbe)",
		Level:   log.ErrorLevel,
	}
	// An AppsV1Deploument should have liveness and readiness endpoints that aren't the same
	APPSV1_DEPLOYMENT_LIVENESS_READINESS_NONMATCHING = &AppsV1DeploymentRule{
		ID:      "APPSV1_DEPLOYMENT_LIVENESS_READINESS_NONMATCHING",
		Prereqs: []RuleID{"V1_PODSPEC_NON_ZERO_CONTAINERS", "APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS", "APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS"},
		Condition: func(deployment *appsv1.Deployment) bool {
			container := deployment.Spec.Template.Spec.Containers[0]
			return container.LivenessProbe.Handler.HTTPGet.Path != container.ReadinessProbe.Handler.HTTPGet.Path
		},
		Message: "It's recommended that the readiness and liveness probe endpoints don't match",
		Level:   log.WarnLevel,
	}
	// A V1PodSpec should have a non-nil security context
	V1_PODSPEC_NON_NIL_SECURITY_CONTEXT = &V1PodSpecRule{
		ID: "V1_PODSPEC_NON_NIL_SECURITY_CONTEXT",
		Condition: func(podSpec *v1.PodSpec) bool {
			return podSpec.SecurityContext != nil
		},
		Message: "The Security context should be present",
		Fix: func(podSpec *v1.PodSpec) bool {
			podSpec.SecurityContext = &corev1.PodSecurityContext{}
			return true
		},
		Level: log.ErrorLevel,
		FixDescription: func(podSpec *v1.PodSpec) string {
			return "Set pod's security context to an empty map"
		},
	}
	// A V1PodSpec should specify runAsNonRoot: true
	V1_PODSPEC_RUN_AS_NON_ROOT = &V1PodSpecRule{
		ID:      "V1_PODSPEC_RUN_AS_NON_ROOT",
		Prereqs: []RuleID{"V1_PODSPEC_NON_NIL_SECURITY_CONTEXT"},
		Condition: func(podSpec *v1.PodSpec) bool {
			return podSpec.SecurityContext.RunAsNonRoot != nil &&
				*podSpec.SecurityContext.RunAsNonRoot == true
		},
		Message: "The Pod Template Spec should enforce that any containers run as non-root users",
		Fix: func(podSpec *v1.PodSpec) bool {
			runAsNonRoot := true
			podSpec.SecurityContext.RunAsNonRoot = &runAsNonRoot
			return true
		},
		Level: log.ErrorLevel,
		FixDescription: func(podSpec *v1.PodSpec) string {
			return "Set pod's runAsNonRoot key to true"
		},
	}
	// A V1PodSpec should have a user and group ID of 44444
	V1_PODSPEC_CORRECT_USER_GROUP_ID = &V1PodSpecRule{
		ID:      "V1_PODSPEC_CORRECT_USER_GROUP_ID",
		Prereqs: []RuleID{"V1_PODSPEC_NON_NIL_SECURITY_CONTEXT"},
		Condition: func(podSpec *v1.PodSpec) bool {
			return podSpec.SecurityContext.RunAsUser != nil &&
				podSpec.SecurityContext.RunAsGroup != nil &&
				*podSpec.SecurityContext.RunAsUser == 44444 &&
				*podSpec.SecurityContext.RunAsGroup == 44444
		},
		Message: "The user and group ID of the podspec should be set to 44444",
		Fix: func(podSpec *v1.PodSpec) bool {
			userId := int64(44444)
			groupId := int64(44444)
			if podSpec.SecurityContext == nil {
				podSpec.SecurityContext = &corev1.PodSecurityContext{}
			}
			podSpec.SecurityContext.RunAsUser = &userId
			podSpec.SecurityContext.RunAsGroup = &groupId
			return true
		},
		Level: log.ErrorLevel,
		FixDescription: func(podSpec *v1.PodSpec) string {
			return "Set pod's User and Group ID to 44444"
		},
	}
	// A V1PodSpec should have exactly one container
	V1_PODSPEC_EXACTLY_1_CONTAINER = &V1PodSpecRule{
		ID: "V1_PODSPEC_EXACTLY_1_CONTAINER",
		Condition: func(podSpec *v1.PodSpec) bool {
			return len(podSpec.Containers) == 1
		},
		Message: "A Podspec should have exactly 1 container",
		Level:   log.ErrorLevel,
	}
	// A V1PodSpec should have a non-zero number of containers
	V1_PODSPEC_NON_ZERO_CONTAINERS = &V1PodSpecRule{
		ID: "V1_PODSPEC_NON_ZERO_CONTAINERS",
		Condition: func(podSpec *v1.PodSpec) bool {
			return len(podSpec.Containers) != 0
		},
		Message: "The Podspec should have at least 1 container defined",
		Level:   log.ErrorLevel,
	}
	// A V1Container should have a non-null security context
	V1_CONTAINER_EXISTS_SECURITY_CONTEXT = &V1ContainerRule{
		ID: "V1_CONTAINER_EXISTS_SECURITY_CONTEXT",
		Condition: func(container *v1.Container) bool {
			return container.SecurityContext != nil
		},
		Message: "A v1/Container should have a non-null security context",
		Level:   log.ErrorLevel,
		Fix: func(container *v1.Container) bool {
			container.SecurityContext = &v1.SecurityContext{}
			return true
		},
		FixDescription: func(container *v1.Container) string {
			return fmt.Sprintf("Set container %s's security context to the empty map", container.Name)
		},
	}
	// A V1Container should not allow privilege escalation
	V1_CONTAINER_ALLOW_PRIVILEGE_ESCALATION_FALSE = &V1ContainerRule{
		ID:      "V1_CONTAINER_ALLOW_PRIVILEGE_ESCALATION_FALSE",
		Prereqs: []RuleID{"V1_CONTAINER_EXISTS_SECURITY_CONTEXT"},
		Condition: func(container *v1.Container) bool {
			return container.SecurityContext.AllowPrivilegeEscalation != nil &&
				*container.SecurityContext.AllowPrivilegeEscalation == false
		},
		Message: "Expected Container's AllowPrivilegeEscalation to be present and set to false",
		Level:   log.ErrorLevel,
		Fix: func(container *v1.Container) bool {
			desired := false
			container.SecurityContext.AllowPrivilegeEscalation = &desired
			return true
		},
		FixDescription: func(container *v1.Container) string {
			return fmt.Sprintf("Set AllowPrivilegeEscalation to false on Container %s", container.Name)
		},
	}
	// A V1Container's image should come from a set of allowed images defined in isImageAllowed
	V1_CONTAINER_VALID_IMAGE = &V1ContainerRule{
		ID: "V1_CONTAINER_VALID_IMAGE",
		Condition: func(container *v1.Container) bool {
			return isImageAllowed(container.Image)
		},
		Message: "The container's image was not from the set of allowed images",
		Level:   log.ErrorLevel,
	}
	// A V1Container should have privileged set to false
	V1_CONTAINER_PRIVILEGED_FALSE = &V1ContainerRule{
		ID:      "V1_CONTAINER_PRIVILEGED_FALSE",
		Prereqs: []RuleID{"V1_CONTAINER_EXISTS_SECURITY_CONTEXT"},
		Condition: func(container *v1.Container) bool {
			return container.SecurityContext.Privileged != nil &&
				*container.SecurityContext.Privileged == false
		},
		Message: "Expected Privileged to be present and set to false",
		Level:   log.ErrorLevel,
		Fix: func(container *v1.Container) bool {
			privileged := false
			container.SecurityContext.Privileged = &privileged
			return true
		},
		FixDescription: func(container *v1.Container) string {
			return fmt.Sprintf("Set Privileged key on container %s to false", container.Name)
		},
	}
	// A V1Container should specify Resource Limits and Requests
	V1_CONTAINER_EXISTS_RESOURCE_LIMITS_AND_REQUESTS = &V1ContainerRule{
		ID: "V1_CONTAINER_EXISTS_RESOURCE_LIMITS_AND_REQUESTS",
		Condition: func(container *v1.Container) bool {
			return container.Resources.Limits != nil && container.Resources.Requests != nil
		},
		Message: "Resource limits must be set for the container (resources.requests) and (resources.limits)",
		Level:   log.ErrorLevel,
	}
	// A V1Container should make CPU requests that are less than or equal to 100%
	V1_CONTAINER_REQUESTS_CPU_REASONABLE = &V1ContainerRule{
		ID:      "V1_CONTAINER_REQUESTS_CPU_REASONABLE",
		Prereqs: []RuleID{"V1_CONTAINER_EXISTS_RESOURCE_LIMITS_AND_REQUESTS"},
		Condition: func(container *v1.Container) bool {
			// If the container is requesting CPU, it shouldn't be more than 1 unit.
			cpuUsage := container.Resources.Requests.Cpu()
			return cpuUsage.CmpInt64(1) != 1
		},
		Message: "You should request less than 1 unit of CPU",
		Level:   log.ErrorLevel,
	}
	// A BatchV1Beta1CronJob should be within a namespace
	BATCHV1_BETA1_CRONJOB_WITHIN_NAMESPACE = &BatchV1Beta1CronJobRule{
		ID: "BATCHV1_BETA1_CRONJOB_WITHIN_NAMESPACE",
		Condition: func(job *batchV1beta1.CronJob) bool {
			return job.Namespace != ""
		},
		Message: "The cronjob must be within a namespace",
		Level:   log.ErrorLevel,
	}
	// A BatchV1Beta1CronJob should forbid concurrent operations
	BATCHV1_BETA1_CRONJOB_FORBID_CONCURRENT = &BatchV1Beta1CronJobRule{
		ID: "BATCHV1_BETA1_CRONJOB_FORBID_CONCURRENT",
		Condition: func(job *batchV1beta1.CronJob) bool {
			return job.Spec.ConcurrencyPolicy == batchV1beta1.ForbidConcurrent
		},
		Message: "Concurent operations should be forbidden",
		Level:   log.ErrorLevel,
		Fix: func(job *batchV1beta1.CronJob) bool {
			job.Spec.ConcurrencyPolicy = batchV1beta1.ForbidConcurrent
			return true
		},
		FixDescription: func(job *batchV1beta1.CronJob) string {
			return fmt.Sprintf("Set concurrency policy on cronjob %s to forbid concurrent", job.Name)
		},
	}

	// A BatchV1Job should be within a namespace
	BATCHV1_JOB_WITHIN_NAMESPACE = &BatchV1JobRule{
		ID: "BATCHV1_JOB_WITHIN_NAMESPACE",
		Condition: func(job *batchV1.Job) bool {
			return job.Namespace != ""
		},
		Message: "A Job should have a namespace specified",
		Level:   log.ErrorLevel,
	}
	// A BatchV1Job's restart policy should be set to Never
	BATCHV1_JOB_RESTART_NEVER = &BatchV1JobRule{
		ID: "BATCHV1_JOB_RESTART_NEVER",
		Condition: func(job *batchV1.Job) bool {
			return len(job.Spec.Template.Spec.RestartPolicy) != 0 &&
				job.Spec.Template.Spec.RestartPolicy == "Never"

		},
		Message: "A Job's restart policy should be set to Never",
		Level:   log.ErrorLevel,
		Fix: func(job *batchV1.Job) bool {
			job.Spec.Template.Spec.RestartPolicy = "Never"
			return true
		},
		FixDescription: func(job *batchV1.Job) string {
			return fmt.Sprintf("Set job %s's restart policy to Never", job.Name)
		},
	}
	// A BatchV1Job's Time to Live should be set
	BATCHV1_JOB_EXISTS_TTL = &BatchV1JobRule{
		ID: "BATCHV1_JOB_EXISTS_TTL",
		Condition: func(job *batchV1.Job) bool {
			return job.Spec.TTLSecondsAfterFinished != nil

		},
		Message: "A Job should set a TTLSecondsAfterFinished value",
		Level:   log.ErrorLevel,
	}
	// A V1Namespace should have a valid DNS name
	V1_NAMESPACE_VALID_DNS = &V1NamespaceRule{
		ID: "V1_NAMESPACE_VALID_DNS",
		Condition: func(namespace *v1.Namespace) bool {
			const ACCEPTABLE_DNS = `^[a-zA-Z][a-zA-Z0-9\-\.]+[a-zA-Z0-9]$`
			validDNS := regexp.MustCompile(ACCEPTABLE_DNS)
			return validDNS.MatchString(namespace.Name)
		},
		Message: "A namespace needs to have a valid DNS name",
		Level:   log.ErrorLevel,
	}
	// A V1Service should be within a namespace
	V1_SERVICE_WITHIN_NAMESPACE = &V1ServiceRule{
		ID: "V1_SERVICE_WITHIN_NAMESPACE",
		Condition: func(service *v1.Service) bool {
			return service.Namespace != ""
		},
		Message: "A service should have a namespace specified",
		Level:   log.ErrorLevel,
	}
	// A V1Service name should be a valid DNS
	V1_SERVICE_NAME_VALID_DNS = &V1ServiceRule{
		ID: "V1_SERVICE_NAME_VALID_DNS",
		Condition: func(service *v1.Service) bool {
			const ACCEPTABLE_DNS = `^[a-zA-Z][a-zA-Z0-9\-\.]+[a-zA-Z0-9]$`
			validDNS := regexp.MustCompile(ACCEPTABLE_DNS)
			return validDNS.MatchString(service.Name)
		},
		Level:   log.ErrorLevel,
		Message: "A service's name needs to be a valid DNS",
	}
)

var (
	// A unit should contain exactly one namespace.
	INTERDEPENDENT_ONE_NAMESPACE = &InterdependentRule{
		ID: "INTERDEPENDENT_ONE_NAMESPACE",
		Condition: func(resources []*Resource) (bool, []*Resource) {
			var namespaces []*Resource
			for _, resource := range resources {
				if resource.TypeInfo.GetKind() == "Namespace" {
					namespaces = append(namespaces, resource)
				}
			}
			return len(namespaces) == 1, namespaces
		},
		Message: "The unit should contain exactly one namespace",
		Level:   log.ErrorLevel,
	}
	// All resources should be under the namespace in the unit
	INTERDEPENDENT_MATCHING_NAMESPACE = &InterdependentRule{
		ID: "INTERDEPENDENT_MATCHING_NAMESPACE",
		Condition: func(resources []*Resource) (bool, []*Resource) {
			var namespace *v1.Namespace
			for _, resource := range resources {
				if ns, ok := resource.Object.(*v1.Namespace); ok {
					namespace = ns
					break
				}
			}
			if namespace == nil {
				return true, nil
			}
			var wrongNamespaceResources []*Resource
			// now test that all ppl are under that namespace
			for _, resource := range resources {
				if resource.TypeInfo.GetKind() == "Namespace" {
					continue
				}
				if resource.Object.GetNamespace() != namespace.Name {
					wrongNamespaceResources = append(wrongNamespaceResources, resource)
				}
			}
			return len(wrongNamespaceResources) == 0, wrongNamespaceResources
		},
		Message: "All resources must be under the correct namespace",
		Level:   log.ErrorLevel,
		Fix: func(resources []*Resource) bool {
			for _, resource := range resources {
				if ns, ok := resource.Object.(*v1.Namespace); ok {
					name := ns.Name
					for _, r := range resources {
						if r.Object == ns {
							continue
						}
						r.Object.SetNamespace(name)
					}
					break
				}
			}
			return true
		},
		FixDescription: func(resources []*Resource) string {
			var namespace string
			for _, resource := range resources {
				if resource.Object.GetNamespace() != "" {
					namespace = resource.Object.GetNamespace()
				}
			}
			return fmt.Sprintf("Set everyone's namespace to %s", namespace)
		},
	}
	// There should be a network policy for the namespace
	INTERDEPENDENT_NETWORK_POLICY_REQUIRED = &InterdependentRule{
		ID: "INTERDEPENDENT_NETWORK_POLICY_REQUIRED",
		Condition: func(resources []*Resource) (bool, []*Resource) {
			var namespaces []*v1.Namespace
			for _, resource := range resources {
				if n, ok := resource.Object.(*v1.Namespace); ok {
					namespaces = append(namespaces, n)
				}
			}
			if len(namespaces) != 1 {
				return true, nil // cuz test isn't relevant
			}
			found := false
			for _, resource := range resources {
				if _, ok := resource.Object.(*networkingV1.NetworkPolicy); ok {
					found = true
					break
				}
			}
			return found, nil
		},
		Message: "There must be a network policy defined",
		Level:   log.ErrorLevel,
	}
)

func isImageAllowed(image string) bool {
	ALLOWED_DOCKER_REGISTRIES := []string{"277433404353.dkr.ecr.eu-central-1.amazonaws.com"}
	for _, r := range ALLOWED_DOCKER_REGISTRIES {
		if strings.HasPrefix(image, r) {
			return true
		}
	}
	return false

}
