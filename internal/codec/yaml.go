package codec

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"strings"
)

var (
	_ Decoder = &yamlDecoder{}
)

func YAMLDecoder(r io.Reader) Decoder {
	d := yaml.NewDecoder(r)
	d.KnownFields(true)
	return &yamlDecoder{d: d}
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

func parseBespoke(field reflect.StructField) map[string]empty {
	bespoke, ok := field.Tag.Lookup("bespoke")
	if !ok {
		return nil
	}
	tagSet := make(map[string]empty)
	for _, tag := range strings.Split(bespoke, ",") {
		key := strings.TrimSpace(tag)
		tagSet[key] = empty{}
	}
	return tagSet
}

func isRequired(field reflect.StructField) bool {
	if set := parseBespoke(field); set != nil {
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
