package codec

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	_ Decoder = &yamlDecoder{}
)

func YAMLCopyTo(to, from any) error {
	buf := new(bytes.Buffer)
	if err := YAMLEncoder(buf).Encode(from); err != nil {
		return err
	}
	d := YAMLDecoder(buf)
	if reflect.TypeOf(to).Kind() != reflect.Pointer {
		return d.Decode(&to)
	}
	return d.Decode(to)
}

// YAMLDecoder parses only known YAML and checks for required fields.
// Required fields are required when they are marked with the
// tag `dinghy:"required"`.
func YAMLDecoder(r io.Reader) Decoder {
	d := yaml.NewDecoder(r)
	d.KnownFields(true)
	return &yamlDecoder{d: d}
}

func YAMLEncoder(w io.Writer) Encoder {
	e := yaml.NewEncoder(w)
	e.SetIndent(2)
	return e
}

type yamlDecoder struct{ d *yaml.Decoder }

func (y *yamlDecoder) Decode(obj any) error {
	if err := y.d.Decode(obj); err != nil {
		return err
	}

	v := reflect.ValueOf(obj).Elem()
	if reflect.TypeOf(v).Kind() != reflect.Struct {
		return nil
	}
	return checkRequired(v)
}

func parseDinghyTag(field reflect.StructField) map[string]empty {
	dinghy, ok := field.Tag.Lookup("dinghy")
	if !ok {
		return nil
	}
	tagSet := make(map[string]empty)
	for _, tag := range strings.Split(dinghy, ",") {
		key := strings.TrimSpace(tag)
		tagSet[key] = empty{}
	}
	return tagSet
}

// isRequired checks a struct field required tags. If a struct
// field is marked as required, isRequired returns true, otherwise
// it returns false
func isRequired(field reflect.StructField) bool {
	if set := parseDinghyTag(field); set != nil {
		_, ok := set["required"]
		return ok
	}
	return false
}

type empty struct{}

func checkRequired(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return nil
	}
	for k := 0; k < v.NumField(); k++ {
		field := v.Type().Field(k)
		name := field.Name
		if tag := field.Tag.Get("yaml"); tag != "" {
			for _, tag := range strings.Split(tag, ",") {
				name = strings.TrimSpace(tag)
				break
			}
		}

		// If the field is required it needs to have a non-zero value
		if isRequired(field) && v.FieldByName(field.Name).IsZero() {
			return ErrRequiredField(name)
		}

		data := v.Field(k)
		switch field.Type.Kind() {
		case reflect.Ptr:
			data = data.Elem()
			fallthrough
		case reflect.Struct:
			if err := checkRequired(data); err != nil {
				req, ok := err.(errRequiredField)
				if !ok {
					return err
				}
				return ErrRequiredField(fmt.Sprintf("%s.%s", name, string(req)))
			}
		case reflect.Slice, reflect.Array:
			n := data.Len()
			for i := 0; i < n; i++ {
				err := checkRequired(data.Index(i))
				if err != nil {
					req, ok := err.(errRequiredField)
					if !ok {
						return err
					}
					return ErrRequiredField(fmt.Sprintf("%s[%d].%s", name, i, string(req)))
				}
			}
		}
	}
	return nil
}
