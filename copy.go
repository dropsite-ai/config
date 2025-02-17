package config

import (
	"errors"
	"fmt"
	"reflect"
)

// CopyProperty copies the value from srcFieldName in srcConfig to dstFieldName in dstConfig.
// Both srcConfig and dstConfig must be pointers to structs.
func CopyProperty(srcConfig interface{}, srcFieldName string, dstConfig interface{}, dstFieldName string) error {
	srcVal := reflect.ValueOf(srcConfig)
	if srcVal.Kind() != reflect.Ptr || srcVal.Elem().Kind() != reflect.Struct {
		return errors.New("srcConfig must be a pointer to a struct")
	}
	dstVal := reflect.ValueOf(dstConfig)
	if dstVal.Kind() != reflect.Ptr || dstVal.Elem().Kind() != reflect.Struct {
		return errors.New("dstConfig must be a pointer to a struct")
	}

	srcField := srcVal.Elem().FieldByName(srcFieldName)
	if !srcField.IsValid() {
		return fmt.Errorf("source field %q not found", srcFieldName)
	}

	dstField := dstVal.Elem().FieldByName(dstFieldName)
	if !dstField.IsValid() {
		return fmt.Errorf("destination field %q not found", dstFieldName)
	}
	if !dstField.CanSet() {
		return fmt.Errorf("cannot set destination field %q", dstFieldName)
	}

	if !srcField.Type().AssignableTo(dstField.Type()) {
		return fmt.Errorf("cannot assign value of type %s to field %q of type %s",
			srcField.Type(), dstFieldName, dstField.Type())
	}

	dstField.Set(srcField)
	return nil
}
