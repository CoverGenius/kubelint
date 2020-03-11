/*
Package kubelint presents a series of utilities to help with keeping your kubernetes resource definitions up to date
and compliant with any semantic rules you want to enforce. One example might be that you want to enforce that all
apps/v1/Deployments have `runAsNonRoot: true`. This might be an easy detail to forget when you are handwriting your resources.
To initialise the linter to enforce this, you just need to set up a linter, add an `AppsV1DeploymentRule` to it, and you
can even implement your own fix. It's up to you what you do with the in-memory representation of your deployment once it is fixed,
but kubelint also provides some utility files to write kubernetes resources to file or as a bytes representation.

This is how to set up the linter to check that every deployment has `runAsNonRoot: true`.

	l := kubelint.NewDefaultLinter()
	l.AddAppsV1DeploymentRule(&kubelint.AppsV1DeploymentRule{
		Condition: func(d *appsv1.Deployment) bool {
			return d.Spec.Template.SecurityContext != nil &&
				   d.Spec.Template.SecurityContext.RunAsNonRoot != nil &&
				  *d.Spec.Template.SecurityContext.RunAsNonRoot == true
		},
		Message: "All deployments should have runAsNonRoot set to true",
		ID: "APPSV1_DEPLOYMENT_RUN_AS_NON_ROOT",
		Level: logrus.ErrorLevel,
	})
	results, errors := l.LintBytes(`kind: Deployment
	version: apps/v1
	metadata:
	  name: hello-world`)
	// If the resources weren't interpretable as YAML kubernetes resources, the errors slice might not be empty
	for _, err := range errors {
		log.Error(err)
	}
	// Log the results of linting!
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	for _, result := range results {
		logger.Log(result.Level, result.Message)
	}
	// If you want the error messages to be more informative, you can pull out information from the `Resources` field
	for _, result := range results {
		logger.WithFields(log.Fields{
			"line number": result.Resources[0].LineNumber,
			"resource kind": result.Resources[0].Resource.TypeInfo.GetKind(),
			"filename": result.Resources[0].Filename,
		}).Log(result.Level, result.Message)
	}

*/
package kubelint
