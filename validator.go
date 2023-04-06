package Validator

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Field string
	Error string
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var b strings.Builder
	if len(v) == 1 {
		b.WriteString(v[0].Error)
	} else {
		for _, e := range v {
			b.WriteString(e.Field + ": " + e.Error + "\n")
		}
	}
	return b.String()
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	var validationErrors ValidationErrors

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		if !field.IsExported() {
			validationErrors = append(validationErrors, ValidationError{
				Field: field.Name,
				Error: ErrValidateForUnexportedFields.Error(),
			})
			continue
		}

		validators := strings.Split(tag, ";")

		for _, validator := range validators {
			parts := strings.SplitN(validator, ":", 2)
			if len(parts) != 2 {
				validationErrors = append(validationErrors, ValidationError{
					Field: field.Name,
					Error: ErrInvalidValidatorSyntax.Error(),
				})
				continue
			}

			switch parts[0] {
			case "len":
				expectedLen, err := strconv.Atoi(parts[1])
				if err != nil {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: ErrInvalidValidatorSyntax.Error(),
					})
					continue
				}
				if fieldValue.Kind() == reflect.String && len(fieldValue.String()) != expectedLen {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: fmt.Sprintf("should have length %d", expectedLen),
					})
				}
			case "in":
				values := strings.Split(parts[1], ",")
				found := false
				empty := false

				for _, value := range values {
					if value == "" {
						empty = true
					}
				}
				if empty {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: ErrInvalidValidatorSyntax.Error(),
					})
					continue
				}

				for _, value := range values {
					if fieldValue.Kind() == reflect.String && fieldValue.String() == value {
						found = true
						break
					} else if fieldValue.Kind() == reflect.Int && strconv.Itoa(int(fieldValue.Int())) == value {
						found = true
						break
					} else if value == "" {
						empty = true
					}
				}
				if !found {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: fmt.Sprintf("should be one of %s", parts[1]),
					})
				}
			case "min":
				expectedMin, err := strconv.Atoi(parts[1])
				if err != nil {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: ErrInvalidValidatorSyntax.Error(),
					})
					continue
				}
				if fieldValue.Kind() == reflect.String && len(fieldValue.String()) < expectedMin {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: fmt.Sprintf("should have length at least %d", expectedMin),
					})
				} else if fieldValue.Kind() == reflect.Int && int(fieldValue.Int()) < expectedMin {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: fmt.Sprintf("should be at least %d", expectedMin),
					})
				}
			case "max":
				expectedMax, err := strconv.Atoi(parts[1])
				if err != nil {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: ErrInvalidValidatorSyntax.Error(),
					})
					continue
				}
				if fieldValue.Kind() == reflect.String && len(fieldValue.String()) > expectedMax {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: fmt.Sprintf("should have length at most %d", expectedMax),
					})
				} else if fieldValue.Kind() == reflect.Int && int(fieldValue.Int()) > expectedMax {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Error: fmt.Sprintf("should be at most %d", expectedMax),
					})
				}
			default:
				validationErrors = append(validationErrors, ValidationError{
					Field: field.Name,
					Error: ErrInvalidValidatorSyntax.Error(),
				})
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}
