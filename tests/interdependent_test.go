package tests

import (
	"testing"

	kubelint "github.com/rdowavic/kubelint-1"
	log "github.com/sirupsen/logrus"
)

func TestInterdependentLinter(t *testing.T) {
	logger := log.New()
	// logger.SetOutput(os.Stdout)
	linter := kubelint.NewLinter(logger)
	linter.AddInterdependentRule(
		kubelint.INTERDEPENDENT_ONE_NAMESPACE,
		kubelint.INTERDEPENDENT_MATCHING_NAMESPACE,
		kubelint.INTERDEPENDENT_NETWORK_POLICY_REQUIRED,
	)
	unit := []byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: pear
---
kind: Namespace
apiVersion: v1
metadata:
  name: mynamespace
`)
	results, errors := linter.LintBytes(unit, "FAKE.yaml")
	for _, err := range errors {
		t.Error(err)
	}
	for _, result := range results {
		t.Logf("%s: %s\n", result.Level, result.Message)
	}
	resources, descriptions := linter.ApplyFixes()
	for _, fix := range descriptions {
		t.Logf("* %s\n", fix)
	}
	bytes, errors := kubelint.Write(resources...)
	for _, err := range errors {
		t.Error(err)
	}
	t.Logf("FIXED:\n%s", bytes)
}
