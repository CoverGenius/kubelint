# kubelint
A Configurable Kubernetes Linter for Go

## Basic Example
The simplest way to use this package is to add predefined rules to a linter object:

```go
package main

import (
    "fmt"
    "github.com/CoverGenius/kubelint"
    "log"
)

func main() {
    linter := kubelint.NewDefaultLinter()
    linter.AddAppsV1DeploymentRule(kubelint.APPSV1_DEPLOYMENT_WITHIN_NAMESPACE)
    linter.AddV1PodSpecRule(kubelint.V1_PODSPEC_RUN_AS_NON_ROOT)
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

```

## Run the linter in debug mode

If you get a runtime panic and don't know what's going on, you can instantiate a logrus linter, set it to debug level,
and pass that to the linter constructor. The reason this happens is because the Rule you've passed in defines a prerequisite that you have forgot to include,
OR you've defined your own rules whose prerequisite specifications are not satisfiable because there is no topological ordering of the rules you've defined (ie, the linter doesn't know which one needs to be evaluated first).

```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
linter := kubelint.NewLinter(logger)
...
```
It will usually make it more obvious which dependent rule is missing that you need to add. 
This will probably be a common trip-up when using predefined rules. To fix the last example, add 
`kubelint.V1_PODSPEC_NON_NIL_SECURITY_CONTEXT`.

```go
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
```
The output is just:
```
The resource must be within a namespace
The Security context should be present
The Pod Template Spec should enforce that any containers run as non-root users
```

# Linter Primitives
The primitives of this package are important to understand before you go ahead and implement your own linter.

### Resources
A resource captures what comes out of a kubernetes YAML definition (an umbrella for all kubernetes object types that are compatible with `meta.Type` and `metav1.Object`). It basically just gives you a `meta.Type` accessor and a `metav1.Object` accessor.
This is here so that you can access common traits, like namespace, name, etc.
```go
result, errs := linter.Lint("example.yaml", "example2.yaml")
resources, fixDescs := linter.ApplyFixes()
for _, resource := range resources {
  fmt.Printf("%s %s fixed!\n", resource.TypeInfo.GetKind(), resource.Object.GetName()) 
}
```
A `YamlDerivedResource` on the other hand is just a resource also, but just augmented with some traits that suggest it was read from a yaml file defined locally.
You can access the filename and linenumber of the `YamlDerivedResource` in addition to the usual accessors you get with the `Resource` type.

```go
ydrs, _ := kubelint.ReadFile(os.Stdin)
for _, ydr := range ydrs {
    fmt.Printf("%s %s on line %d in file %s\n", 
      ydr.Resource.TypeInfo.GetKind(),
      ydr.Resource.Object.GetName(),
      ydr.LineNumber,
      ydr.Filepath,
    )
}
```
You will have to work with the `Resource` type if you want your linter to apply fixes and print them to stdout or save them to disk.

### Results

Once a rule in your linter is evaluated on a specific object, a `Result` is created and returned to you from the `Lint` invocation. The intention is for you to be able to report
to the user what went wrong, and how bad it was. Each result also keeps track of the offending `YamlDerivedResource`, so that you know which resource needs fixing.

```go
logger := logrus.New()
results, errs := linter.LintFile(os.Stdin)
// results is type []*Result
for _, result := range results {
    logger.Logf(result.Level, "%s %s: %s", 
      result.Resources[0].TypeInfo.GetKind(),
      result.Resources[0].Object.GetName(),
      result.Message,
    )
}
```
I recommend using a `logrus.Logger` so that you can pass in the `result.Level` field and have the log coloured in the suitable way!

### The Linter Object
All a linter does is store a bunch of rules. When you invoke the linter with a `Lint` function, 
the files or filepaths that you pass in are unmarshalled and stored within the linter. The linter then iterates through all the rules
that you've assigned to it, and matches resources with rules that correspond in type. You primarily feed it files or filepaths or alternatively bytes,
and get back a list of results that you can log.

### Rules
The linter keeps track of type-specific rules (eg `RbacV1Beta1RoleBindingRule`, `NetworkingV1NetworkPolicyRule`), generic rules (`GenericRule`), and interdependent rules (`InterdependentRule`) 
(that require a scan over every resource in order to evaluate). A rule should capture some kind of semantic requirement 
on your kubernetes objects. An example is that every deployment name should contain the string "coronavirus". This should be defined
by the `Condition func(*appsv1.Deployment) bool` field.

```go
import (
  "github.com/CoverGenius/kubelint"
  appsv1 "k8s.io/api/apps/v1"
)
func main() {
  myRule := &kubelint.AppsV1DeploymentRule{
    Condition: func(d *appsv1.Deployment) bool {
                return strings.Contains(d.Name, "coronavirus")
      },
    }
}
```
When the condition isn't satisfied, ideally you'd want to let the user know. Define the string to give back when the rule fails by setting `Message`.

```
myRule := &kubelint.AppsV1DeploymentRule{
  Condition: func(d *appsv1.Deployment) bool {
                return strings.contains(d.Name, "coronavirus")
             },
  Message: "The deployment name doesn't contain coronavirus, and I am viewing this as a dire problem",
  Level: logrus.ErrorLevel,
  
}
```
Make sure you define `Level`. If not, it will default to `PanicLevel`.

#### Prerequisites
Sometimes, it helps to be able to factor rules. For example, you need to check the length of a slice field (`ID: "IMPORTANT_LENGTH_CHECK"`) before you 
check the contents of the slice (`ID: FIRST_CONTAINER_IS_CORONA_FREE`). It might feel painful to perform the nil-check over and over again, so you can factor this out into its own rule, and then any rule that relies on this one to evaluate successfully should have `Prereqs: []RuleID{"IMPORTANT_LENGTH_CHECK"}`.

```go
rule := &kubelint.AppsV1DeploymentRule{
  ID: "FIRST_CONTAINER_CORONA_FREE",
  Prereqs: []RuleID{"IMPORTANT_LENGTH_CHECK"},
  Condition: func(d *appsv1.Deployment) bool {
    return d.Spec.Template.Spec.Containers[0].Name != "Corona"
  },
}
```
Then you can ensure that the dereference `[0]` won't cause a runtime panic because the `"IMPORTANT_LENGTH_CHECK"` must have been evaluated first, and was successful.

#### Fixes
You can attach a Fix method to all of your rules! It's expected that this just mutates the object in some way so that the rule is satisfied. You signal that the rule has been satisfied by returning `true`, and `false` if it wasn't possible to fix the object.

```go
rule := &kubelint.AppsV1DeploymentRule{
  ID: "FIRST_CONTAINER_CORONA_FREE",
  Prereqs: []RuleID{"IMPORTANT_LENGTH_CHECK"},
  Condition: func(d *appsv1.Deployment) bool {
    return d.Spec.Template.Spec.Containers[0].Name != "Corona"
  },
  Level: logrus.ErrorLevel,
  Fix: func(d *appsv1.Deployment) bool {
    d.Spec.Template.Spec.Containers[0].Name = "Corona"
    return true
  },
  FixDescription: func(d *appsv1.Deployment) string {
     return fmt.Sprintf("Set Deployment %s's name to Corona", d.Name)
  }
}
```

### Interdependent Rules
Sometimes, you can't actually evaluate if a condition is met by looking at resources one by one. You need to judge the collection of resources as a whole.
For example, everything you lint should be under the namespace that you are also linting. If the namespace is missing, you'd like to apply an automatic fix to have the namespace changed to the correct namespace.
This is an example of when you should add an interdependent rule.

### Unsupported Types
Ideally, just fork this repo and add a `AddMyFavouriteTypeRule` method and an extra field to the linter to store rules of this type.
You will also need to implement a conversion function from `MyFavouriteType -> rule`, an unexported type that is just the result of interpolating the concrete object into the `Condition` body, etc.
It should be clear from the existing examples under `rule.go`. Otherwise, what you can do is create a `GenericRule` (and this also applies if you want to apply the same check across all types).
It is exactly the same as a type-specific rule, except the type is `*Resource` rather than `*appsv1.Deployment`, for example. You can do your own concrete typecast within the body of the function. 
Just note that this rule WILL be applied to every single object that you send in for linting.





