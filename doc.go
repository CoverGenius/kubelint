/*
Package kubelint presents a series of utilities to help with keeping your kubernetes resource definitions up to date
and compliant with any semantic rules you want to enforce. One example might be that you want to enforce that all
apps/v1/Deployments have runAsNonRoot: true. This might be an easy detail to forget when you are handwriting your resources.
To initialise the linter to enforce this, you just need to set up a linter, add an AppsV1DeploymentRule to it, and you
can even implement your own fix. It's up to you what you do with the in-memory representation of your deployment once it is fixed,
but kubelint also provides some utility functions to write kubernetes resources to file or as a bytes representation.

To summarise, the main objectives you can fulfill with kubelint are:

1. Linting YAML kubernetes resource for correctness

2. Reporting resource definition errors

3. Automatically applying fixes to resource definitions

4. Writing kubernetes resources to a file

5. Obtaining the bytes representation of a fixed resource definition

This is how to set up the linter to check that every deployment has runAsNonRoot: true.

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
Then once you have a linter set up  with some rules, all it takes to have the linter perform checks is to provide a reference
to the filepath or a bytes slice.
	results, errors := l.LintBytes([]byte(`kind: Deployment
	version: apps/v1
	metadata:
	  name: hello-world`))
	// If the resources weren't interpretable as YAML kubernetes resources, the errors slice might not be empty
	for _, err := range errors {
		logrus.Error(err)
	}
	// Log the results of linting!
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	for _, result := range results {
		logger.Log(result.Level, result.Message)
	}

The result might look like:
	ERRO[0000] All deployments should have runAsNonRoot set to true  filepath=example_yamls/deployment_non_root_false.yaml line number=1 resource name=hello-world-web

If you want the error messages to be more informative, you can pull out information from the Resources field. For most results,
there will only be one resource in this slice, but for rules referencing many resources, it could be multiple. You would need to
define those yourself.

	for _, result := range results {
		logger.WithFields(log.Fields{
			"line number": result.Resources[0].LineNumber,
			"resource kind": result.Resources[0].Resource.TypeInfo.GetKind(),
			"filename": result.Resources[0].Filename,
		}).Log(result.Level, result.Message)
	}
You can also use any of the predefined rules in the kubelint package to pass to your linter.
Just make sure that if any prerequisites are listed within the rule definition that you include these too.
The prerequisites are rules that must be executed before the current rule. This is usually because the prerequisite
rule tests for a nil field or performs a length check necessary for a dereference operation. You don't have to use this feature
if you don't see its use case, but it might come in handy if you need to factor your rules down the track.
The rules are evaluated in topological sort order to make sure that prerequisite rules are executed before any dependent rules.

	filepaths := []string{"my_favourite_deployment.yaml", "really_insecure_deployment.yaml"}
	l := kubelint.NewDefaultLinter()
	l.AddAppsV1DeploymentRule(
		kubelint.APPSV1_DEPLOYMENT_WITHIN_NAMESPACE,
		kubelint.APPSV1_DEPLOYMENT_CONTAINER_EXISTS_LIVENESS,
		kubelint.APPSV1_DEPLOYMENT_CONTAINER_EXISTS_READINESS,
	)
	results, errors := l.Lint(filepaths...)

If you want to apply fixes, you can write the result to whatever file you want, or just output the result to a bytes slice.
You can't modify the file in-place so to speak, since we perform the analysis on in-memory representations of the contents of
the original files, but you can still simulate an in-place fixer if you overwrite the original file with the result of l.ApplyFixes().

	l := kubelint.NewDefaultLinter()
	// ... add some rules
	results, errors := l.Lint(filepaths...)
	// you can apply the fixes that are suggested
	resources, fixDescriptions := l.ApplyFixes()
	bytes, errs := kubelint.Write(resources...)
	fmt.Printf("%s\n", string(bytes))

Also notice that you can report the fixes that have been applied.
	for _, description := range fixDescriptions {
		fmt.Printf("X %s\n", description)
	}
If anything goes wrong with your linter implementation, you can attempt to debug by creating a logrus.logger instance
and passing it to the constructor of the linter object.

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	linter := kubelint.NewLinter(logger)
	...
*/
package kubelint
