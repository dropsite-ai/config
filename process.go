package config

import (
	"reflect"
	"strings"
)

// Process recursively processes cfg (struct pointer, map, slice, etc.)
// for fields or keys ending with Secret, Path, User, or URL.
func Process(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	if !val.IsValid() {
		return nil
	}
	return processValue(val, "")
}

func processValue(v reflect.Value, name string) error {
	// 1) Unwrap interface until we get to the concrete value
	for v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil
	}

	// 2) If pointer, dereference and recurse
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return processValue(v.Elem(), name)
	}

	switch v.Kind() {

	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			fieldVal := v.Field(i)
			fieldType := t.Field(i)

			// Skip unexported fields
			if fieldType.PkgPath != "" {
				continue
			}
			fieldName := fieldType.Name

			if fieldVal.Kind() == reflect.String && fieldVal.CanSet() {
				if err := processStringField(fieldVal, fieldName); err != nil {
					return err
				}
			} else {
				if err := processValue(fieldVal, fieldName); err != nil {
					return err
				}
			}
		}
		return nil

	case reflect.Map:
		// Expect map[string]interface{} or map[string]T
		if v.Type().Key().Kind() != reflect.String {
			return nil
		}
		for _, key := range v.MapKeys() {
			k := key.String()
			valItem := v.MapIndex(key)

			// Unwrap interface
			for valItem.Kind() == reflect.Interface && !valItem.IsNil() {
				valItem = valItem.Elem()
			}
			if !valItem.IsValid() {
				continue
			}

			switch valItem.Kind() {
			case reflect.String:
				updated, err := processString(valItem.String(), k)
				if err != nil {
					return err
				}
				v.SetMapIndex(key, reflect.ValueOf(updated))
			case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array, reflect.Ptr, reflect.Interface:
				// Recurse. But we typically do so on a copy, then store it back.
				newVal := reflect.New(valItem.Type()).Elem()
				newVal.Set(valItem)
				if err := processValue(newVal, k); err != nil {
					return err
				}
				v.SetMapIndex(key, newVal)
			default:
				// No special suffix logic for other kinds
			}
		}
		return nil

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			elemVal := v.Index(i)
			// Unwrap interface
			for elemVal.Kind() == reflect.Interface && !elemVal.IsNil() {
				elemVal = elemVal.Elem()
			}
			if !elemVal.IsValid() {
				continue
			}

			// If it's a string, we can process directly and store result.
			if elemVal.Kind() == reflect.String && elemVal.CanSet() {
				updated, err := processString(elemVal.String(), name)
				if err != nil {
					return err
				}
				v.Index(i).Set(reflect.ValueOf(updated))
				continue
			}

			// Otherwise, handle nested maps/structs by recursion
			if elemVal.Kind() == reflect.Map ||
				elemVal.Kind() == reflect.Struct ||
				elemVal.Kind() == reflect.Slice ||
				elemVal.Kind() == reflect.Array ||
				elemVal.Kind() == reflect.Ptr ||
				elemVal.Kind() == reflect.Interface {

				newVal := reflect.New(elemVal.Type()).Elem()
				newVal.Set(elemVal)
				if err := processValue(newVal, name); err != nil {
					return err
				}
				// Put updated item back into slice
				v.Index(i).Set(newVal)
			}
		}
		return nil

	case reflect.String:
		// Possibly do expansions if it ends with Secret, Path, etc.
		if v.CanSet() {
			updated, err := processString(v.String(), name)
			if err != nil {
				return err
			}
			v.SetString(updated)
		}
		return nil
	}

	// For all other types, do nothing
	return nil
}

func processStringField(fieldVal reflect.Value, fieldName string) error {
	current := fieldVal.String()
	updated, err := processString(current, fieldName)
	if err != nil {
		return err
	}
	fieldVal.SetString(updated)
	return nil
}

// processString applies Secret/Path/User/URL expansions based on field/key name.
func processString(s string, fieldName string) (string, error) {
	switch {
	case strings.HasSuffix(fieldName, "Secret"):
		if s == "" {
			newSecret, err := GenerateJWTSecret()
			if err != nil {
				return s, err
			}
			return newSecret, nil
		}
		return s, nil

	case strings.HasSuffix(fieldName, "Path"):
		if s != "" {
			expanded, err := ExpandPath(s)
			if err != nil {
				return s, err
			}
			return expanded, nil
		}
		return s, nil

	case strings.HasSuffix(fieldName, "User"):
		if s != "" {
			if err := ValidateUsername(s); err != nil {
				return s, err
			}
		}
		return s, nil

	case strings.HasSuffix(fieldName, "URL"):
		if s != "" {
			if err := ValidateURL(s); err != nil {
				return s, err
			}
		}
		return s, nil

	default:
		// No special rules
		return s, nil
	}
}
