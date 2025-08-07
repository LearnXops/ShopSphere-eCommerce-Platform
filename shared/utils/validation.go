package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/shopspring/decimal"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface for ValidationError
func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string, value interface{}) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Validator provides validation functionality
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Errors returns the validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// Required validates that a field is not empty
func (v *Validator) Required(field string, value interface{}) *Validator {
	if isEmpty(value) {
		v.errors.Add(field, "is required", value)
	}
	return v
}

// Email validates email format
func (v *Validator) Email(field, email string) *Validator {
	if email == "" {
		return v
	}
	
	if _, err := mail.ParseAddress(email); err != nil {
		v.errors.Add(field, "must be a valid email address", email)
	}
	return v
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.errors.Add(field, fmt.Sprintf("must be at least %d characters long", min), value)
	}
	return v
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if len(value) > max {
		v.errors.Add(field, fmt.Sprintf("must be at most %d characters long", max), value)
	}
	return v
}

// Range validates that a number is within a range
func (v *Validator) Range(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.errors.Add(field, fmt.Sprintf("must be between %d and %d", min, max), value)
	}
	return v
}

// DecimalRange validates that a decimal is within a range
func (v *Validator) DecimalRange(field string, value, min, max decimal.Decimal) *Validator {
	if value.LessThan(min) || value.GreaterThan(max) {
		v.errors.Add(field, fmt.Sprintf("must be between %s and %s", min.String(), max.String()), value)
	}
	return v
}

// Positive validates that a number is positive
func (v *Validator) Positive(field string, value int) *Validator {
	if value <= 0 {
		v.errors.Add(field, "must be positive", value)
	}
	return v
}

// DecimalPositive validates that a decimal is positive
func (v *Validator) DecimalPositive(field string, value decimal.Decimal) *Validator {
	if value.LessThanOrEqual(decimal.Zero) {
		v.errors.Add(field, "must be positive", value)
	}
	return v
}

// Pattern validates that a string matches a regex pattern
func (v *Validator) Pattern(field, value, pattern, message string) *Validator {
	if value == "" {
		return v
	}
	
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.errors.Add(field, message, value)
	}
	return v
}

// Phone validates phone number format (basic validation)
func (v *Validator) Phone(field, phone string) *Validator {
	if phone == "" {
		return v
	}
	
	// Remove common separators
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")
	
	// Basic validation: should be 10-15 digits, optionally starting with +
	pattern := `^\+?[1-9]\d{9,14}$`
	if matched, _ := regexp.MatchString(pattern, cleaned); !matched {
		v.errors.Add(field, "must be a valid phone number", phone)
	}
	return v
}

// Username validates username format
func (v *Validator) Username(field, username string) *Validator {
	if username == "" {
		return v
	}
	
	// Username should be 3-30 characters, alphanumeric and underscores only
	if len(username) < 3 || len(username) > 30 {
		v.errors.Add(field, "must be between 3 and 30 characters", username)
		return v
	}
	
	pattern := `^[a-zA-Z0-9_]+$`
	if matched, _ := regexp.MatchString(pattern, username); !matched {
		v.errors.Add(field, "can only contain letters, numbers, and underscores", username)
	}
	return v
}

// Password validates password strength
func (v *Validator) Password(field, password string) *Validator {
	if password == "" {
		return v
	}
	
	if len(password) < 8 {
		v.errors.Add(field, "must be at least 8 characters long", nil)
		return v
	}
	
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		v.errors.Add(field, "must contain at least one uppercase letter", nil)
	}
	if !hasLower {
		v.errors.Add(field, "must contain at least one lowercase letter", nil)
	}
	if !hasDigit {
		v.errors.Add(field, "must contain at least one digit", nil)
	}
	if !hasSpecial {
		v.errors.Add(field, "must contain at least one special character", nil)
	}
	
	return v
}

// SKU validates SKU format
func (v *Validator) SKU(field, sku string) *Validator {
	if sku == "" {
		return v
	}
	
	// SKU should be 3-50 characters, alphanumeric and hyphens only
	if len(sku) < 3 || len(sku) > 50 {
		v.errors.Add(field, "must be between 3 and 50 characters", sku)
		return v
	}
	
	pattern := `^[A-Z0-9-]+$`
	if matched, _ := regexp.MatchString(pattern, sku); !matched {
		v.errors.Add(field, "can only contain uppercase letters, numbers, and hyphens", sku)
	}
	return v
}

// Rating validates rating value (1-5)
func (v *Validator) Rating(field string, rating int) *Validator {
	if rating < 1 || rating > 5 {
		v.errors.Add(field, "must be between 1 and 5", rating)
	}
	return v
}

// OneOf validates that a value is one of the allowed values
func (v *Validator) OneOf(field string, value interface{}, allowed []interface{}) *Validator {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return v
		}
	}
	
	v.errors.Add(field, fmt.Sprintf("must be one of: %v", allowed), value)
	return v
}

// Custom allows custom validation with a function
func (v *Validator) Custom(field string, value interface{}, fn func(interface{}) error) *Validator {
	if err := fn(value); err != nil {
		v.errors.Add(field, err.Error(), value)
	}
	return v
}

// isEmpty checks if a value is empty
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case bool:
		return !v
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// SanitizeString removes potentially harmful characters from a string
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Remove control characters except newlines and tabs
	var result strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			continue
		}
		result.WriteRune(r)
	}
	
	return result.String()
}

// SanitizeHTML removes HTML tags from a string (basic implementation)
func SanitizeHTML(input string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// ValidateUserRegistration validates user registration data
func ValidateUserRegistration(email, username, firstName, lastName, password string) ValidationErrors {
	v := NewValidator()
	
	v.Required("email", email).Email("email", email)
	v.Required("username", username).Username("username", username)
	v.Required("first_name", firstName).MaxLength("first_name", firstName, 50)
	v.Required("last_name", lastName).MaxLength("last_name", lastName, 50)
	v.Required("password", password).Password("password", password)
	
	return v.Errors()
}

// ValidateProductCreation validates product creation data
func ValidateProductCreation(sku, name, description string, price decimal.Decimal) ValidationErrors {
	v := NewValidator()
	
	v.Required("sku", sku).SKU("sku", sku)
	v.Required("name", name).MaxLength("name", name, 200)
	v.Required("description", description).MaxLength("description", description, 2000)
	v.Required("price", price).DecimalPositive("price", price)
	
	return v.Errors()
}

// ValidateReviewCreation validates review creation data
func ValidateReviewCreation(rating int, title, content string) ValidationErrors {
	v := NewValidator()
	
	v.Required("rating", rating).Rating("rating", rating)
	v.Required("title", title).MaxLength("title", title, 100)
	v.Required("content", content).MaxLength("content", content, 2000)
	
	return v.Errors()
}