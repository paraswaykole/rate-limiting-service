package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
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

		value := v.Field(i)
		if (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) &&
			value.Type().Elem().Kind() == reflect.Int64 {
			b, _ := json.Marshal(value.Interface())
			m[tag] = b
		} else {
			m[tag] = value.Interface()
		}
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
				// Handle time.Time specifically
				if v.Field(i).Type() == reflect.TypeOf(time.Time{}) {
					// Try parsing using RFC3339
					if parsedTime, err := time.Parse(time.RFC3339, val); err == nil {
						v.Field(i).Set(reflect.ValueOf(parsedTime))
					}
					continue
				}

				// Handle slices/arrays of structs from JSON
				if (v.Field(i).Kind() == reflect.Slice || v.Field(i).Kind() == reflect.Array) &&
					v.Field(i).Type().Elem().Kind() == reflect.Int {
					slicePtr := reflect.New(v.Field(i).Type()).Interface()
					if err := json.Unmarshal([]byte(val), slicePtr); err == nil {
						v.Field(i).Set(reflect.ValueOf(slicePtr).Elem())
					}
					continue
				}
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

func SendValidationErrors(err error, c fiber.Ctx) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"field": e.Field(),
				"error": e.Error(),
			})
		}
	}
}

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano()) // Seed RNG
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func GetFloat64Slice(m map[string]any, key string) ([]float64, error) {
	raw, exists := m[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found", key)
	}

	switch v := raw.(type) {
	case []float64:
		return v, nil
	case []any:
		out := make([]float64, len(v))
		for i, val := range v {
			num, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("value at index %d is not a float64", i)
			}
			out[i] = num
		}
		return out, nil
	case json.RawMessage:
		var out []float64
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, fmt.Errorf("failed to unmarshal RawMessage: %v", err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unexpected type %T for key %s", raw, key)
	}
}
