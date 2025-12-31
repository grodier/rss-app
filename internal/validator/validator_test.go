package validator

import (
	"regexp"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("expected NewValidator to return a non-nil validator")
	}

	if v.Errors == nil {
		t.Fatal("expected Errors map to be initialized")
	}

	if len(v.Errors) != 0 {
		t.Errorf("expected new validator to have no errors, got %d", len(v.Errors))
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Validator)
		expected bool
	}{
		{
			name:     "new validator is valid",
			setup:    func(v *Validator) {},
			expected: true,
		},
		{
			name: "validator with errors is invalid",
			setup: func(v *Validator) {
				v.Errors["field"] = "error message"
			},
			expected: false,
		},
		{
			name: "validator with multiple errors is invalid",
			setup: func(v *Validator) {
				v.Errors["field1"] = "error 1"
				v.Errors["field2"] = "error 2"
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			tt.setup(v)

			if v.Valid() != tt.expected {
				t.Errorf("expected Valid() to be %v, got %v", tt.expected, v.Valid())
			}
		})
	}
}

func TestAddError(t *testing.T) {
	t.Run("adds error for new key", func(t *testing.T) {
		v := NewValidator()
		v.AddError("email", "invalid email")

		if len(v.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(v.Errors))
		}

		if v.Errors["email"] != "invalid email" {
			t.Errorf("expected error message 'invalid email', got '%s'", v.Errors["email"])
		}
	})

	t.Run("does not overwrite existing error", func(t *testing.T) {
		v := NewValidator()
		v.AddError("email", "first error")
		v.AddError("email", "second error")

		if len(v.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(v.Errors))
		}

		if v.Errors["email"] != "first error" {
			t.Errorf("expected error message 'first error', got '%s'", v.Errors["email"])
		}
	})

	t.Run("can add errors for different keys", func(t *testing.T) {
		v := NewValidator()
		v.AddError("email", "invalid email")
		v.AddError("password", "password too short")

		if len(v.Errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(v.Errors))
		}

		if v.Errors["email"] != "invalid email" {
			t.Errorf("expected email error, got '%s'", v.Errors["email"])
		}

		if v.Errors["password"] != "password too short" {
			t.Errorf("expected password error, got '%s'", v.Errors["password"])
		}
	})
}

func TestCheck(t *testing.T) {
	t.Run("adds error when condition is false", func(t *testing.T) {
		v := NewValidator()
		v.Check(false, "age", "must be positive")

		if len(v.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(v.Errors))
		}

		if v.Errors["age"] != "must be positive" {
			t.Errorf("expected error message 'must be positive', got '%s'", v.Errors["age"])
		}
	})

	t.Run("does not add error when condition is true", func(t *testing.T) {
		v := NewValidator()
		v.Check(true, "age", "must be positive")

		if len(v.Errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(v.Errors))
		}
	})

	t.Run("respects AddError behavior for duplicate keys", func(t *testing.T) {
		v := NewValidator()
		v.Check(false, "email", "first error")
		v.Check(false, "email", "second error")

		if len(v.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(v.Errors))
		}

		if v.Errors["email"] != "first error" {
			t.Errorf("expected first error to be kept, got '%s'", v.Errors["email"])
		}
	})
}

func TestPermittedValue(t *testing.T) {
	t.Run("returns true for permitted string value", func(t *testing.T) {
		result := PermittedValue("red", "red", "blue", "green")

		if !result {
			t.Error("expected true for permitted value")
		}
	})

	t.Run("returns false for non-permitted string value", func(t *testing.T) {
		result := PermittedValue("yellow", "red", "blue", "green")

		if result {
			t.Error("expected false for non-permitted value")
		}
	})

	t.Run("works with integers", func(t *testing.T) {
		if !PermittedValue(5, 1, 3, 5, 7) {
			t.Error("expected true for permitted integer")
		}

		if PermittedValue(4, 1, 3, 5, 7) {
			t.Error("expected false for non-permitted integer")
		}
	})

	t.Run("works with empty permitted values", func(t *testing.T) {
		result := PermittedValue("test")

		if result {
			t.Error("expected false when no permitted values provided")
		}
	})

	t.Run("works with single permitted value", func(t *testing.T) {
		if !PermittedValue("admin", "admin") {
			t.Error("expected true for single matching value")
		}

		if PermittedValue("user", "admin") {
			t.Error("expected false for single non-matching value")
		}
	})
}

func TestMatches(t *testing.T) {
	t.Run("returns true for matching pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{4}$`)
		result := Matches("123-4567", pattern)

		if !result {
			t.Error("expected true for matching pattern")
		}
	})

	t.Run("returns false for non-matching pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{4}$`)
		result := Matches("123-45678", pattern)

		if result {
			t.Error("expected false for non-matching pattern")
		}
	})

	t.Run("validates email with EmailRX", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"test.user@example.com",
			"user+tag@example.co.uk",
		}

		for _, email := range validEmails {
			if !Matches(email, EmailRX) {
				t.Errorf("expected '%s' to be a valid email", email)
			}
		}
	})

	t.Run("rejects invalid emails with EmailRX", func(t *testing.T) {
		invalidEmails := []string{
			"notanemail",
			"@example.com",
			"user@",
			"user @example.com",
		}

		for _, email := range invalidEmails {
			if Matches(email, EmailRX) {
				t.Errorf("expected '%s' to be an invalid email", email)
			}
		}
	})
}

func TestUnique(t *testing.T) {
	t.Run("returns true for unique integers", func(t *testing.T) {
		values := []int{1, 2, 3, 4, 5}
		result := Unique(values)

		if !result {
			t.Error("expected true for unique values")
		}
	})

	t.Run("returns false for duplicate integers", func(t *testing.T) {
		values := []int{1, 2, 3, 2, 5}
		result := Unique(values)

		if result {
			t.Error("expected false for duplicate values")
		}
	})

	t.Run("returns true for unique strings", func(t *testing.T) {
		values := []string{"apple", "banana", "cherry"}
		result := Unique(values)

		if !result {
			t.Error("expected true for unique values")
		}
	})

	t.Run("returns false for duplicate strings", func(t *testing.T) {
		values := []string{"apple", "banana", "apple"}
		result := Unique(values)

		if result {
			t.Error("expected false for duplicate values")
		}
	})

	t.Run("returns true for empty slice", func(t *testing.T) {
		values := []int{}
		result := Unique(values)

		if !result {
			t.Error("expected true for empty slice")
		}
	})

	t.Run("returns true for single element slice", func(t *testing.T) {
		values := []string{"only"}
		result := Unique(values)

		if !result {
			t.Error("expected true for single element slice")
		}
	})
}

func TestValidatorWorkflow(t *testing.T) {
	t.Run("validates user input with multiple checks", func(t *testing.T) {
		v := NewValidator()

		username := "john"
		email := "invalid-email"
		age := 15

		v.Check(len(username) >= 5, "username", "must be at least 5 characters")
		v.Check(Matches(email, EmailRX), "email", "must be a valid email")
		v.Check(age >= 18, "age", "must be at least 18")

		if v.Valid() {
			t.Error("expected validator to be invalid")
		}

		expectedErrors := 3
		if len(v.Errors) != expectedErrors {
			t.Errorf("expected %d errors, got %d", expectedErrors, len(v.Errors))
		}
	})

	t.Run("validates permitted values", func(t *testing.T) {
		v := NewValidator()

		role := "superadmin"
		v.Check(PermittedValue(role, "user", "admin"), "role", "must be user or admin")

		if v.Valid() {
			t.Error("expected validator to be invalid")
		}

		if v.Errors["role"] != "must be user or admin" {
			t.Errorf("expected role error, got '%s'", v.Errors["role"])
		}
	})

	t.Run("validates unique values", func(t *testing.T) {
		v := NewValidator()

		tags := []string{"go", "testing", "go"}
		v.Check(Unique(tags), "tags", "must not contain duplicates")

		if v.Valid() {
			t.Error("expected validator to be invalid")
		}

		if v.Errors["tags"] != "must not contain duplicates" {
			t.Errorf("expected tags error, got '%s'", v.Errors["tags"])
		}
	})
}
