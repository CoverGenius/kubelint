package main

import (
	"os"

	"github.com/CoverGenius/kubelint"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

func main() {
	linter := kubelint.NewDefaultLinter()
	// linter configuration
	linter.AddAppsV1DeploymentRule(&kubelint.AppsV1DeploymentRule{
		Condition: func(d *appsv1.Deployment) bool {
			return d.Spec.Template.Spec.SecurityContext != nil &&
				d.Spec.Template.Spec.SecurityContext.RunAsNonRoot != nil &&
				*d.Spec.Template.Spec.SecurityContext.RunAsNonRoot == true
		},
		Message: "All deployments should have runAsNonRoot set to true",
		ID:      "APPSV1_DEPLOYMENT_NON_ROOT",
		Level:   log.ErrorLevel,
	})
	results, errs := linter.Lint("example_yamls/deployment_non_root_false.yaml")
	for _, err := range errs {
		log.Error(err)
	}
	logger := log.StandardLogger()
	logger.SetOutput(os.Stdout)
	for _, result := range results {
		logger.WithFields(log.Fields{
			"line number":   result.Resources[0].LineNumber,
			"filepath":      result.Resources[0].Filepath,
			"resource name": result.Resources[0].Resource.Object.GetName(),
		}).Log(result.Level, result.Message)
	}
}
