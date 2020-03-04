package tests

import (
	"fmt"
	"strings"

	"github.com/rdowavic/kubelint"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"

	"testing"
)

func TestLinterBasic(t *testing.T) {
	linter := kubelint.NewDefaultLinter()
	linter.AddAppsV1DeploymentRule(
		&kubelint.AppsV1DeploymentRule{
			ID: "DEPLOYMENT_NAME_CONTAINS_APPLE",
			Condition: func(d *appsv1.Deployment) bool {
				return strings.Contains(d.Name, "apple")
			},
			Message: "A deployment's name needs to contain the string \"apple\"",
			Level:   log.ErrorLevel,
			Fix: func(d *appsv1.Deployment) bool {
				d.Name += "-apple" // fantastic fix
				return true
			},
			FixDescription: func(d *appsv1.Deployment) string {
				return fmt.Sprintf("Changed deployment's name to %s", d.Name)
			},
		})
	results, errs := linter.LintBytes([]byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: pear
`), "FAKE_DEPLOYMENT.yaml")
	for _, result := range results {
		t.Log(result.Message)
	}
	t.Log(len(errs))
	for _, err := range errs {
		t.Error(err)
	}
	resources, fixDescriptions := linter.ApplyFixes()
	for _, resource := range resources {
		bytes, _ := kubelint.Write(resource)
		t.Log(string(bytes))
		if !strings.Contains(resource.Object.GetName(), "apple") {
			t.Errorf("Deployment's name was not successfully changed to end with apple")
		}
	}
	for _, desc := range fixDescriptions {
		t.Logf("* %s\n", desc)
	}
}

func TestLinterPredefinedRule(t *testing.T) {
	linter := kubelint.NewDefaultLinter()
	linter.AddAppsV1DeploymentRule(kubelint.APPSV1_DEPLOYMENT_WITHIN_NAMESPACE)
	d := []byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: doggy
`)
	results, errs := linter.LintBytes(d, "FAKE_DEPLOYMENT.yaml")
	for _, result := range results {
		t.Log(result.Message)
	}
	for _, err := range errs {
		t.Error(err)
	}
}

func TestLinterInterdependentRule(t *testing.T) {
	linter := kubelint.NewDefaultLinter()
	linter.AddInterdependentRule(&kubelint.InterdependentRule{
		ID:      "INTERDEPENDENT_NO_DUPLICATES",
		Message: "The unit shouldn't contain duplicate resources (same name/kind)",
		Condition: func(resources []*kubelint.Resource) bool {
			nameKind := make(map[string]string) // a set storing <name><kind>
			for _, resource := range resources {
				nk := resource.Object.GetName() + resource.TypeInfo.GetKind()
				if _, exists := nameKind[nk]; exists {
					return false
				}
				nameKind[nk] = nk
			}
			return true
		},
	})
	bytes := []byte(`kind: Service
apiVersion: v1
metadata:
  name: barista
---
kind: Service
apiVersion: v1
metadata:
  name: barista
`)
	results, errs := linter.LintBytes(bytes, "FAKE.yaml")

	for _, result := range results {
		t.Log(result.Message)
	}
	for _, err := range errs {
		t.Error(err)
	}

}
