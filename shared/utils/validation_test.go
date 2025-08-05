package utils

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidator_Required(t *testing.T) {
	v := NewValidator()
	
	// Test empty string
	v.Required("field1", "")
	if !v.HasErrors() {
		t.Error("Expected validation error for empty string")
	}
	
	// Test non-empty string
	v2 := NewValidator()
	v2.Required("field1", "value")
	if v2.HasErrors() {
		t.Error("Expected no validation error for non-empty string")
	}
}

func TestValidator_Email(t *testing.T) {
	v := NewValidator()
	
	// Test valid email
	v.Email("email", "test@example.com")
	if v.HasErrors() {
		t.Error("Expected no validation error for valid email")
	}
	
	// Test invalid email
	v2 := NewValidator()
	v2.Email("email", "invalid-email")
	if !v2.HasErrors() {
		t.Error("Expected validation error for invalid email")
	}
}

func TestValidator_Password(t *testing.T) {
	v := NewValidator()
	
	// Test weak password
	v.Password("password", "weak")
	if !v.HasErrors() {
		t.Error("Expected validation error for weak password")
	}
	
	// Test strong password
	v2 := NewValidator()
	v2.Password("password", "StrongP@ss123")
	if v2.HasErrors() {
		t.Errorf("Expected no validation error for strong password, got: %v", v2.Errors())
	}
}

func TestValidator_Rating(t *testing.T) {
	v := NewValidator()
	
	// Test invalid rating
	v.Rating("rating", 6)
	if !v.HasErrors() {
		t.Error("Expected validation error for rating > 5")
	}
	
	// Test valid rating
	v2 := NewValidator()
	v2.Rating("rating", 4)
	if v2.HasErrors() {
		t.Error("Expected no validation error for valid rating")
	}
}

func TestValidateUserRegistration(t *testing.T) {
	// Test valid registration
	errors := ValidateUserRegistration("test@example.com", "testuser", "John", "Doe", "StrongP@ss123")
	if errors.HasErrors() {
		t.Errorf("Expected no validation errors for valid registration, got: %v", errors)
	}
	
	// Test invalid registration
	errors2 := ValidateUserRegistration("", "", "", "", "")
	if !errors2.HasErrors() {
		t.Error("Expected validation errors for empty registration data")
	}
}

func TestValidateProductCreation(t *testing.T) {
	price := decimal.NewFromFloat(19.99)
	
	// Test valid product
	errors := ValidateProductCreation("PROD-123", "Test Product", "A test product", price)
	if errors.HasErrors() {
		t.Errorf("Expected no validation errors for valid product, got: %v", errors)
	}
	
	// Test invalid product
	errors2 := ValidateProductCreation("", "", "", decimal.Zero)
	if !errors2.HasErrors() {
		t.Error("Expected validation errors for empty product data")
	}
}

func TestSanitizeString(t *testing.T) {
	input := "  Hello\x00World\t  "
	expected := "HelloWorld" // null bytes and control chars removed, whitespace trimmed
	result := SanitizeString(input)
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSanitizeHTML(t *testing.T) {
	input := "<script>alert('xss')</script>Hello <b>World</b>"
	expected := "alert('xss')Hello World" // HTML tags removed
	result := SanitizeHTML(input)
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}