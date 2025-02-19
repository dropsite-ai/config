package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// CopyProperty now supports nested fields using dot-notation, e.g. "Nested.Field".
// Both srcConfig and dstConfig must be pointers to structs.
func CopyProperty(srcConfig interface{}, srcFieldPath string, dstConfig interface{}, dstFieldPath string) error {
	srcVal := reflect.ValueOf(srcConfig)
	if srcVal.Kind() != reflect.Ptr || srcVal.Elem().Kind() != reflect.Struct {
		return errors.New("srcConfig must be a pointer to a struct")
	}
	dstVal := reflect.ValueOf(dstConfig)
	if dstVal.Kind() != reflect.Ptr || dstVal.Elem().Kind() != reflect.Struct {
		return errors.New("dstConfig must be a pointer to a struct")
	}

	srcField, err := getNestedField(srcVal.Elem(), srcFieldPath)
	if err != nil {
		return err
	}
	dstField, err := getNestedFieldForWrite(dstVal.Elem(), dstFieldPath)
	if err != nil {
		return err
	}
	if !srcField.Type().AssignableTo(dstField.Type()) {
		return fmt.Errorf("cannot assign value of type %s to field %q of type %s",
			srcField.Type(), dstFieldPath, dstField.Type())
	}
	dstField.Set(srcField)
	return nil
}

// getNestedField traverses the nested fields specified by a dot-separated path (for reading).
func getNestedField(val reflect.Value, path string) (reflect.Value, error) {
	parts := strings.Split(path, ".")
	for _, part := range parts {
		// Dereference pointers if necessary.
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return reflect.Value{}, fmt.Errorf("nil pointer encountered in path %q", path)
			}
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("expected struct while processing %q in path %q", part, path)
		}
		field := val.FieldByName(part)
		if !field.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %q not found in path %q", part, path)
		}
		val = field
	}
	return val, nil
}

// getNestedFieldForWrite traverses the nested fields and returns the final field,
// ensuring that it is settable.
func getNestedFieldForWrite(val reflect.Value, path string) (reflect.Value, error) {
	parts := strings.Split(path, ".")
	for i, part := range parts {
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return reflect.Value{}, fmt.Errorf("nil pointer encountered in path %q", path)
			}
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("expected struct while processing %q in path %q", part, path)
		}
		// For the final part, ensure the field is settable.
		if i == len(parts)-1 {
			field := val.FieldByName(part)
			if !field.IsValid() {
				return reflect.Value{}, fmt.Errorf("field %q not found in path %q", part, path)
			}
			if !field.CanSet() {
				return reflect.Value{}, fmt.Errorf("cannot set field %q in path %q", part, path)
			}
			return field, nil
		}
		// Otherwise, traverse to the next nested struct.
		val = val.FieldByName(part)
		if !val.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %q not found in path %q", part, path)
		}
	}
	return val, nil
}
