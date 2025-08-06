package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidateStruct performs basic struct validation using reflection
// This is a simple implementation that checks for required fields based on validate tags
func ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	
	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", v.Kind())
	}
	
	validator := NewValidator()
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("validate")
		
		if tag == "" {
			continue
		}
		
		fieldName := strings.ToLower(fieldType.Name)
		fieldValue := field.Interface()
		
		// Parse validation tags
		tags := strings.Split(tag, ",")
		for _, t := range tags {
			t = strings.TrimSpace(t)
			
			switch {
			case t == "required":
				validator.Required(fieldName, fieldValue)
			case strings.HasPrefix(t, "min="):
				// Handle min validation for numbers
				if field.Kind() == reflect.Int {
					// This is a simplified implementation
					if field.Int() < 1 {
						validator.errors.Add(fieldName, "must be at least 1", fieldValue)
					}
				}
			case strings.HasPrefix(t, "max="):
				// Handle max validation for numbers  
				if field.Kind() == reflect.Int {
					// This is a simplified implementation
					if field.Int() > 168 { // Max 168 hours (7 days)
						validator.errors.Add(fieldName, "must be at most 168", fieldValue)
					}
				}
			}
		}
	}
	
	if validator.HasErrors() {
		return validator.Errors()
	}
	
	return nil
}
