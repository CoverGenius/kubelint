package main

import (
	"os"

	"github.com/rdowavic/kubelint"
	log "github.com/sirupsen/logrus"
)

func main() {
	linter := kubelint.NewDefaultLinter()
	linter.AddV1PodSpecRule(kubelint.V1_PODSPEC_CORRECT_USER_GROUP_ID, kubelint.V1_PODSPEC_NON_NIL_SECURITY_CONTEXT)
	results, errs := linter.Lint("example_yamls/deployment_invalid_user_group_ids.yaml")
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
