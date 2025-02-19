package config

import (
	"reflect"
)

// Process looks for a "Variables" section (or a "variables" key if cfg is a map)
// and then applies processing to its sub-maps: endpoints, secrets, users, and paths.
func Process(cfg interface{}) error {
	if cfg == nil {
		return nil
	}
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// If cfg is a struct, check for a field named "Variables".
	if v.Kind() == reflect.Struct {
		field := v.FieldByName("Variables")
		if field.IsValid() {
			return processVariables(field)
		}
	}

	// If cfg is a map, look for a key "variables".
	if v.Kind() == reflect.Map {
		key := reflect.ValueOf("variables")
		val := v.MapIndex(key)
		if val.IsValid() {
			// Process the "variables" value.
			if err := processVariables(val); err != nil {
				return err
			}
			// (Optional) reassign the processed value back to the map.
			v.SetMapIndex(key, val)
			return nil
		}
	}

	return nil
}

// processVariables expects v to be a struct or map containing the keys (or fields)
// for endpoints, secrets, users, and paths. It applies the appropriate processing
// for each.
func processVariables(v reflect.Value) error {
	// Dereference pointer if needed.
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Process by field names (expecting "Endpoints", "Secrets", "Users", "Paths").
		if field := v.FieldByName("Endpoints"); field.IsValid() && field.Kind() == reflect.Map {
			if err := processMap(field, processEndpointValue); err != nil {
				return err
			}
		}
		if field := v.FieldByName("Secrets"); field.IsValid() && field.Kind() == reflect.Map {
			if err := processMap(field, processSecretValue); err != nil {
				return err
			}
		}
		if field := v.FieldByName("Users"); field.IsValid() && field.Kind() == reflect.Map {
			if err := processMap(field, processUserValue); err != nil {
				return err
			}
		}
		if field := v.FieldByName("Paths"); field.IsValid() && field.Kind() == reflect.Map {
			if err := processMap(field, processPathValue); err != nil {
				return err
			}
		}
	case reflect.Map:
		// Process by looking for keys "endpoints", "secrets", "users", and "paths".
		if endpoints := v.MapIndex(reflect.ValueOf("endpoints")); endpoints.IsValid() {
			if err := processMapValue(endpoints, processEndpointValue); err != nil {
				return err
			}
		}
		if secrets := v.MapIndex(reflect.ValueOf("secrets")); secrets.IsValid() {
			if err := processMapValue(secrets, processSecretValue); err != nil {
				return err
			}
		}
		if users := v.MapIndex(reflect.ValueOf("users")); users.IsValid() {
			if err := processMapValue(users, processUserValue); err != nil {
				return err
			}
		}
		if paths := v.MapIndex(reflect.ValueOf("paths")); paths.IsValid() {
			if err := processMapValue(paths, processPathValue); err != nil {
				return err
			}
		}
	}

	return nil
}

// processMap iterates over a map (as a struct field) and updates each string value
// using the provided processor function.
func processMap(m reflect.Value, processor func(string) (string, error)) error {
	for _, key := range m.MapKeys() {
		val := m.MapIndex(key)
		var s string
		if val.Kind() == reflect.String {
			s = val.String()
		} else if val.CanInterface() {
			if str, ok := val.Interface().(string); ok {
				s = str
			} else {
				continue
			}
		}
		newS, err := processor(s)
		if err != nil {
			return err
		}
		m.SetMapIndex(key, reflect.ValueOf(newS))
	}
	return nil
}

// processMapValue is similar to processMap but works on map values obtained from a map.
func processMapValue(v reflect.Value, processor func(string) (string, error)) error {
	if v.Kind() != reflect.Map {
		return nil
	}
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		var s string
		if val.Kind() == reflect.String {
			s = val.String()
		} else if val.CanInterface() {
			if str, ok := val.Interface().(string); ok {
				s = str
			} else {
				continue
			}
		}
		newS, err := processor(s)
		if err != nil {
			return err
		}
		v.SetMapIndex(key, reflect.ValueOf(newS))
	}
	return nil
}

// processEndpointValue validates that the string is a valid URL.
func processEndpointValue(s string) (string, error) {
	if s == "" {
		return s, nil
	}
	if err := ValidateURL(s); err != nil {
		return s, err
	}
	return s, nil
}

// processSecretValue generates a new secret if the value is empty.
func processSecretValue(s string) (string, error) {
	if s == "" {
		return GenerateJWTSecret()
	}
	return s, nil
}

// processUserValue validates the username.
func processUserValue(s string) (string, error) {
	if s == "" {
		return s, nil
	}
	if err := ValidateUsername(s); err != nil {
		return s, err
	}
	return s, nil
}

// processPathValue expands the path (e.g. replacing "~" with the home directory).
func processPathValue(s string) (string, error) {
	if s == "" {
		return s, nil
	}
	return ExpandPath(s)
}
