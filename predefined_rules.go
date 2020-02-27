package kubelint

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

/**
* Predefined rules relating to resources of type appsv1.Deployment
* An AppsV1Deployment should have a project label *	APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL
* An AppsV1Deployment should have an app.kubernetes.io/name label *  APPSV1_DEPLOYMENT_EXISTS_APP_K8S_LABEL
* An AppsV1Deployment should be within a namespace * APPSV1_DEPLOYMENT_WITHIN_NAMESPACE
*
*
*
**/

const (
	// An AppsV1Deployment should have a project label.
	APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL = &AppsV1DeploymentRule{
		ID: "APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL",
		Condition: func(deployment *appsv1.Deployment) bool {
			_, found := deployment.Spec.Template.Labels["project"]
			return found
		},
		Message: "There should be a project label present under the deployment's spec.template.labels",
		Level:   log.ErrorLevel,
		Fix:     func(*appsv1.Deployment) bool { return false },
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
		Fix:     func(*appsv1.Deployment) bool { return false },
	}
	// An AppsV1Deployment should specify a liveness endpoint
	APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS = &AppsV1DeploymentRule{
		ID:      "APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS",
		Prereqs: []RuleID{PODSPEC_NON_ZERO_CONTAINERS},
		Condition: func(deployment *appsv1.Deployment) bool {
			return deployment.Spec.Template.Spec.Containers[0].LivenessProbe != nil &&
				deployment.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.HTTPGet != nil
		},
		Message: "Expected declaration of liveness probe for the container (livenessProbe)",
		Level:   log.ErrorLevel,
		Fix:     func(deployment *appsv1.Deployment) bool { return false },
	}
	// An AppsV1Deployment should specify a readiness endpoint
	APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS = &AppsV1DeploymentRule{
		ID:      "APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS",
		Prereqs: []RuleID{PODSPEC_NON_ZERO_CONTAINERS},
		Condition: func(deployment *appsv1.Deployment) bool {
			return deployment.Spec.Template.Spec.Containers[0].ReadinessProbe != nil &&
				deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.HTTPGet != nil
		},
		Message: "Expected declaration of readiness probe for the container (readinessProbe)",
		Level:   log.ErrorLevel,
		Fix:     func(deployment *appsv1.Deployment) bool { return false },
	}
	// An AppsV1Deploument should have liveness and readiness endpoints that aren't the same
	APPSV1_DEPLOYMENT_LIVENESS_READINESS_NONMATCHING = &AppsV1DeploymentRule{
		ID:      "APPSV1_DEPLOYMENT_LIVENESS_READINESS_NONMATCHING",
		Prereqs: []RuleID{PODSPEC_NON_ZERO_CONTAINERS, DEPLOYMENT_CONTAINER_EXISTS_READINESS, DEPLOYMENT_CONTAINER_EXISTS_LIVENESS},
		Condition: func(deployment *appsv1.Deployment) bool {
			container := deployment.Spec.Template.Spec.Containers[0]
			return container.LivenessProbe.Handler.HTTPGet.Path != container.ReadinessProbe.Handler.HTTPGet.Path
		},
		Message: "It's recommended that the readiness and liveness probe endpoints don't match",
		Level:   log.WarnLevel,
		Fix:     func(deployment *appsv1.Deployment) bool { return false },
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
		Level:          log.ErrorLevel,
		FixDescription: fmt.Sprintf("Set resource %s's security context to an empty map", outerResourceName),
	}
	// A V1PodSpec should specify runAsNonRoot: true	
	V1_PODSPEC_RUN_AS_NON_ROOT = &V1PodSpecRule{
		ID:      "V1_PODSPEC_RUN_AS_NON_ROOT",
		Prereqs: []RuleID{V1_PODSPEC_NON_NIL_SECURITY_CONTEXT},
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
		Level:          log.ErrorLevel,
		Resources:      []*YamlDerivedKubernetesResource{resource},
		FixDescription: fmt.Sprintf("Set runAsNonRoot to true for resource %s", outerResourceName),
	}
	V1_PODSPEC_CORRECT_USER_GROUP_ID = &V1PodSpecRule{
			ID:      "V1_PODSPEC_CORRECT_USER_GROUP_ID",
			Prereqs: []RuleID{V1_PODSPEC_NON_NIL_SECURITY_CONTEXT},
			Condition: func(podSpec *v1.PodSpec) bool {
				return podSpec.SecurityContext.RunAsUser != nil &&
					podSpec.SecurityContext.RunAsGroup != nil &&
					*podSpec.SecurityContext.RunAsUser == 44444 &&
					*podSpec.SecurityContext.RunAsGroup == 44444
			},
			Message: "The user and group ID should be set to 44444",
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
			Level:          log.ErrorLevel,
			Resources:      []*YamlDerivedKubernetesResource{resource},
			FixDescription: fmt.Sprintf("Set resource %s's user ID and group ID to 44444", outerResourceName),
	}
	V1_PODSPEC_CORRECT_USER_GROUP_ID = &V1PodSpecRule{

			ID: V1_PODSPEC_EXACTLY_1_CONTAINER,
			Condition: func(podSpec *v1.PodSpec) bool {
				return len(podSpec.Containers) == 1
			},
			Message:   fmt.Sprintf("There should be exactly 1 container defined for %s, but there are %#v", outerResourceName, len(podSpec.Containers)),
			Level:     log.ErrorLevel,
			Resources: []*YamlDerivedKubernetesResource{resource},
			Fix: func(podSpec *v1.PodSpec) bool {
				return false // not possible to fix this by myself!  which one would they keep?
			},
		},
		{
			ID: V1_PODSPEC_NON_ZERO_CONTAINERS,
			Condition: func(podSpec *v1.PodSpec) bool {
				return len(podSpec.Containers) != 0
			},
			Message:   fmt.Sprintf("%s should have at least 1 container defined", outerResourceName),
			Resources: []*YamlDerivedKubernetesResource{resource},
			Fix: func(podSpec *v1.PodSpec) bool {
				return false
			},

)
