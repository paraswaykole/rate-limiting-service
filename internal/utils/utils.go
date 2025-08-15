package utils

import (
	"errors"
	"reflect"
	"strconv"
)

func StructToMap(data any) map[string]any {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// If pointer, get the element
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	m := make(map[string]any)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get JSON tag if present
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		m[tag] = v.Field(i).Interface()
	}

	return m
}

func MapToStruct(m map[string]string, out any) error {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("out must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		if val, ok := m[tag]; ok {
			// Set only if field can be set
			if v.Field(i).CanSet() {
				// Handle basic types
				switch v.Field(i).Kind() {
				case reflect.String:
					v.Field(i).SetString(val)
				case reflect.Float64:
					// Parse float
					if f, err := strconv.ParseFloat(val, 64); err == nil {
						v.Field(i).SetFloat(f)
					}
				case reflect.Int, reflect.Int64:
					if iv, err := strconv.ParseInt(val, 10, 64); err == nil {
						v.Field(i).SetInt(iv)
					}
					// You can add more types like time.Time parsing if needed
				}
			}
		}
	}

	return nil
}
