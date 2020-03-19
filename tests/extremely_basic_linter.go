package main

import (
	"fmt"
	kubelint "github.com/rdowavic/kubelint-1"
	"log"
)

func main() {
	linter := kubelint.NewDefaultLinter()
	linter.AddAppsV1DeploymentRule(kubelint.APPSV1_DEPLOYMENT_WITHIN_NAMESPACE)
	linter.AddV1PodSpecRule(kubelint.V1_PODSPEC_RUN_AS_NON_ROOT, kubelint.V1_PODSPEC_NON_NIL_SECURITY_CONTEXT)
	results, errs := linter.LintBytes([]byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: hello-world
`), "fake_file.yaml")
	for _, err := range errs {
		log.Println(err)
	}

	for _, result := range results {
		fmt.Println(result.Message)
	}
}
