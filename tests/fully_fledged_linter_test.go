package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/fatih/color"
	kubelint "github.com/rdowavic/kubelint-1"
	log "github.com/sirupsen/logrus"
)

var (
	Fix    bool = true
	Report bool = true
)

func TestFullyFledgedLinter(t *testing.T) {
	// Prepare the linter.
	l := log.New()
	l.SetOutput(os.Stdout)
	linter := kubelint.NewLinter(l)
	linter.AddAppsV1DeploymentRule(
		kubelint.APPSV1_DEPLOYMENT_EXISTS_PROJECT_LABEL,
		kubelint.APPSV1_DEPLOYMENT_EXISTS_APP_K8S_LABEL,
		kubelint.APPSV1_DEPLOYMENT_WITHIN_NAMESPACE,
		kubelint.APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS,
		kubelint.APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS,
		kubelint.APPSV1_DEPLOYMENT_LIVENESS_READINESS_NONMATCHING,
	)
	linter.AddV1PodSpecRule(
		kubelint.V1_PODSPEC_NON_NIL_SECURITY_CONTEXT,
		kubelint.V1_PODSPEC_RUN_AS_NON_ROOT,
		kubelint.V1_PODSPEC_CORRECT_USER_GROUP_ID,
		kubelint.V1_PODSPEC_EXACTLY_1_CONTAINER,
		kubelint.V1_PODSPEC_NON_ZERO_CONTAINERS,
	)
	linter.AddV1ContainerRule(
		kubelint.V1_CONTAINER_EXISTS_SECURITY_CONTEXT,
		kubelint.V1_CONTAINER_ALLOW_PRIVILEGE_ESCALATION_FALSE,
		kubelint.V1_CONTAINER_VALID_IMAGE,
		kubelint.V1_CONTAINER_PRIVILEGED_FALSE,
		kubelint.V1_CONTAINER_EXISTS_RESOURCE_LIMITS_AND_REQUESTS,
		kubelint.V1_CONTAINER_REQUESTS_CPU_REASONABLE,
	)
	linter.AddBatchV1Beta1CronJobRule(
		kubelint.BATCHV1_BETA1_CRONJOB_WITHIN_NAMESPACE,
		kubelint.BATCHV1_BETA1_CRONJOB_FORBID_CONCURRENT,
	)
	linter.AddBatchV1JobRule(
		kubelint.BATCHV1_JOB_WITHIN_NAMESPACE,
		kubelint.BATCHV1_JOB_RESTART_NEVER,
		kubelint.BATCHV1_JOB_EXISTS_TTL,
	)
	linter.AddV1NamespaceRule(
		kubelint.V1_NAMESPACE_VALID_DNS,
	)
	linter.AddV1ServiceRule(
		kubelint.V1_SERVICE_NAME_VALID_DNS,
		kubelint.V1_SERVICE_WITHIN_NAMESPACE,
		kubelint.V1_SERVICE_NAME_VALID_DNS,
	)
	filepaths := []string{"../examples/example_yamls/gc/deployment_invalid_user_group_ids.yaml", "../examples/example_yamls/deployment_messed_up.yaml", "../examples/example_yamls/deployment_missing_security_context_privilege.yaml", "../examples/example_yamls/deployment_wrong_privilege_escalation.yaml", "../examples/example_yamls/invalid_job.yaml", "../examples/example_yamls/partially_wrong_unit_directory/Deployment.yaml", "../examples/example_yamls/partially_wrong_unit_directory/Namespace.yaml", "../examples/example_yamls/partially_wrong_unit_directory/NetworkPolicy.yaml", "../examples/example_yamls/partially_wrong_unit_directory/NetworkPolicy1.yaml", "../examples/example_yamls/partially_wrong_unit_directory/Role.yaml", "../examples/example_yamls/partially_wrong_unit_directory/RoleBinding.yaml", "../examples/example_yamls/partially_wrong_unit_directory/Service.yaml", "../examples/example_yamls/partially_wrong_unit_directory/ServiceAccount.yaml", "../examples/example_yamls/valid_cronjob.yaml", "../examples/example_yamls/valid_deployment.yaml", "../examples/example_yamls/valid_job.yaml", "../examples/example_yamls/valid_namespace.yaml", "../examples/example_yamls/valid_unit.yaml"}

	results, errors := linter.Lint(filepaths...)
	logger := log.New()
	for _, err := range errors {
		logger.Error(err)
	}
	logger.SetOutput(os.Stdout)
	for _, result := range results {
		logger.WithFields(log.Fields{
			"line number":   result.Resources[0].LineNumber,
			"filepath":      result.Resources[0].Filepath,
			"resource name": result.Resources[0].Resource.Object.GetName(),
		}).Log(result.Level, result.Message)
	}

	// write out the report if they want it!
	if Fix {
		resources, fixDescriptions := linter.ApplyFixes()
		byteRepresentation, errs := kubelint.Write(resources...)
		if len(errs) != 0 {
			for err := range errs {
				log.Error(err)
			}
			os.Exit(1)
		}
		// output to stdout by default
		fmt.Printf(string(byteRepresentation))
		if Report {
			ReportFixes(fixDescriptions)
		}
	}
}

func ReportFixes(errorFixes []string) {
	green := color.New(color.FgHiGreen).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(os.Stderr, "=====%s=====\n", bold("FIX SUMMARY"))
	for _, errorFix := range errorFixes {
		fmt.Fprintf(os.Stderr, " %s %s\n", green("✓"), errorFix)
	}
}
