package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "builtin.dinghy.dev"

// ErrPatchStrategicMergeUnregisteredSchema occurs when a provided resource
// doesn't have a struct type registered scheme.Scheme. The struct should
// contain the tags that identify fields to merge on, such as the `name`
// field of a container
type ErrPatchStrategicMergeUnregisteredSchema struct {
	GroupVersionKind schema.GroupVersionKind
	Name             string
	Namespace        string
	Patch            map[string]any
}

func (e *ErrPatchStrategicMergeUnregisteredSchema) Error() string {
	name := e.Name
	if e.Namespace != "" {
		name = e.Namespace + "/" + name
	}
	return fmt.Sprintf(`
No registered schema for patch on resource %[3]s with kind %[2]s.

A strategicMergePatch requires a schema that contains merge keys used 
to identify merge points of a resource, for example, the "name" attribute
of a container identifies a container in an array of containers. Without
the schema, a strategicMergePatch is no different than a merge patch.
Please consider using one of the following mutators instead

* %[1]s/patch - merge a single value using an array of field paths as a target 
   Examples
   --------
   uses: %[1]s/patch
   selector:
     kinds:
     - apps/v1/Deployment
   with:
     fieldPaths:
     - spec.template.spec.containers[0]
     value:
       imagePullPolicy: Never
       image: nginx:latest
     merge: true

* %[1]s/scripts/js - use inline javascript to mutate the resource
   Examples
   --------
   uses: %[1]s/patch
   selector:
     kinds:
     - apps/v1/Deployment
   with:
     script: |
       function mutate(obj, c) {
         obj.metadata.name == c.namePrefix + obj.metadata.name
         obj.metadata.namespace = c.namespace
       }
     config:
       namePrefix: "foo-"
       namespace: "bar"
`, GroupName, e.GroupVersionKind, name)
}

// ErrPatchStrategicMerge occurs whiles applying the strategicMergePatch to
// a resource.
type ErrPatchStrategicMerge struct {
	schema.GroupVersionKind
	Name      string
	Namespace string
	Err       error
	Patch     map[string]any
	Resource  map[string]any
}

func (e ErrPatchStrategicMerge) Error() string {
	name := e.Name
	if e.Namespace != "" {
		name = e.Namespace + "/" + name
	}
	buf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(e.Resource); err != nil {

	}

	objBytes := buf.Bytes()

	buf.Reset()
	encoder = yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(e.Patch); err != nil {

	}

	patchBytes := buf.Bytes()

	return fmt.Sprintf(`
Failed to apply strategicMergePatch to resource %[1]s

Resource:
  %[2]s

Patch: 
  %[3]s

Error caused by: %[4]s
`, name, string(objBytes), string(patchBytes), e.Err)
}

// List is a list of errors with a few helpful functions
type List struct {
	errs []error
}

func (l *List) Error() string {
	errs := make([]string, 0, len(l.errs))
	for _, err := range l.errs {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "\n\n")
}

func (l *List) Empty() bool {
	return len(l.errs) == 0
}

func (l *List) First() error {
	if l.Empty() {
		return nil
	}
	return l.errs[0]
}

func (l *List) Append(err error) {
	if in, ok := err.(*List); ok {
		l.Extend(in)
		return
	}
	l.errs = append(l.errs, err)
}

func (l *List) Extend(in *List) {
	for _, e := range in.errs {
		l.Append(e)
	}
}

type ErrDecodeGenerator struct {
	Name   string
	Schema *jsonschema.Schema
	Err    error
}

func (err ErrDecodeGenerator) Error() string {
	data, _ := json.MarshalIndent(err.Schema, "", "  ")
	return fmt.Sprintf(`
The generator %q could not be decoded into it's typed config

Schema: %s

Caused by: %s
`, err.Name, string(data), fmt.Sprintf(err.Err.Error()))
}

type ErrParseSourcePath struct {
	Path string
	Err  error
}

func (err ErrParseSourcePath) Error() string {
	// TODO: get more info on available paths
	return fmt.Sprintf(`
Failed to parse the source path "%q"

Caused by: %s
`, err.Path, err.Err)
}

// ErrMergePatch occurs whiles applying the mergePatch to a resource.
type ErrMergePatch struct {
	schema.GroupVersionKind
	Name      string
	Namespace string
	Err       error
	Patch     map[string]any
	Resource  map[string]any
}

func (e ErrMergePatch) Error() string {
	name := e.Name
	if e.Namespace != "" {
		name = e.Namespace + "/" + name
	}
	buf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(e.Resource); err != nil {
	}

	objBytes := buf.Bytes()

	buf.Reset()
	encoder = yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(e.Patch); err != nil {
	}

	patchBytes := buf.Bytes()

	return fmt.Sprintf(`
Failed to apply mergepatch to resource %[1]s

Resource:
  %[2]s

Patch: 
  %[3]s

Error caused by: %[4]s
`, name, string(objBytes), string(patchBytes), e.Err)
}
