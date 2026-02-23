package web

import (
	"encoding"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

type ParseError struct {
	Message string
	Field   string
	Err     error
}

func (p *ParseError) Error() string {
	return fmt.Sprintf("%s field %s: %v", p.Message, p.Field, p.Err)
}

func (p *ParseError) Is(err error) bool {
	return errors.Is(p.Err, err)
}

func (p *ParseError) As(target any) bool {
	return errors.As(p.Err, target)
}

func handleBool(subFieldVal reflect.Value, valStr string) *ParseError {
	// For checkbox, missing value or not "on"/"true" means false
	if valStr == "on" || valStr == "true" {
		subFieldVal.SetBool(true)
	} else {
		subFieldVal.SetBool(false)
	}
	return nil
}

func handleInt(subFieldVal reflect.Value, valStr string) *ParseError {
	if valStr == "" {
		valStr = "0"
	}
	n, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return &ParseError{Message: "invalid integer", Err: err}
	}
	subFieldVal.SetInt(n)
	return nil
}

func handleUint(subFieldVal reflect.Value, valStr string) *ParseError {
	if valStr == "" {
		valStr = "0"
	}
	n, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return &ParseError{Message: "invalid integer", Err: err}
	}
	subFieldVal.SetUint(n)
	return nil
}

func handleFloat(subFieldVal reflect.Value, valStr string) *ParseError {
	if valStr == "" {
		valStr = "0"
	}
	n, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return &ParseError{Message: "invalid integer", Err: err}
	}
	subFieldVal.SetFloat(n)
	return nil
}

func handleField(subFieldVal reflect.Value, valStr string) *ParseError {
	// Handle boolean/checkbox fields
	if subFieldVal.Kind() == reflect.Bool {
		return handleBool(subFieldVal, valStr)
	}

	// Handle TextUnmarshaler
	if subFieldVal.CanAddr() {
		if tu, ok := subFieldVal.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := tu.UnmarshalText([]byte(valStr)); err != nil {
				return &ParseError{Message: "failed to unmarshal", Err: err}
			}
			return nil
		}
	}

	// Standard types
	switch subFieldVal.Kind() {
	case reflect.String:
		subFieldVal.SetString(valStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return handleInt(subFieldVal, valStr)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return handleUint(subFieldVal, valStr)
	case reflect.Float32, reflect.Float64:
		return handleFloat(subFieldVal, valStr)
	}

	return nil
}

func (p *configPage[T]) updateConfig(sectionName string, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	v := reflect.ValueOf(p.config).Elem()
	sectionField := v.FieldByName(sectionName)
	if !sectionField.IsValid() || sectionField.Kind() != reflect.Struct {
		return fmt.Errorf("section %s not found", sectionName)
	}

	st := sectionField.Type()
	for i := 0; i < sectionField.NumField(); i++ {
		subField := st.Field(i)
		subFieldVal := sectionField.Field(i)

		if subField.PkgPath != "" {
			continue
		}

		// Determine field name used in form (default to struct field name, override by tag)
		field := parseTag(subFieldVal, subField)

		valStr := r.FormValue(field.Name)

		if err := handleField(subFieldVal, valStr); err != nil {
			err.Field = field.Name
			return err
		}
	}
	if sectionField.CanAddr() {
		if ur, ok := sectionField.Addr().Interface().(UpdateReceiver); ok {
			return ur.Updated(p.config, p)
		}
	}
	return nil
}
