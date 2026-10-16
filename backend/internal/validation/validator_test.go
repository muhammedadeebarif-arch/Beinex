package validation

import (
	"testing"

	"ems-backend/internal/model"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Trims whitespace", "  john.doe@example.com  ", "john.doe@example.com"},
		{"Converts to lowercase", "John.Doe@EXAMPLE.COM", "john.doe@example.com"},
		{"Mixed case and spaces", "  User.Name+Tag@Domain.CO.UK  ", "user.name+tag@domain.co.uk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeEmail(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeEmail(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateCreateEmployee(t *testing.T) {
	validReq := func() model.CreateEmployeeRequest {
		return model.CreateEmployeeRequest{
			Name:        "Alice Johnson",
			Email:       "alice@beinex.com",
			Department:  "Engineering",
			Designation: "Software Engineer",
			JoiningDate: "2024-01-15",
			Status:      "Active",
		}
	}

	t.Run("Valid request produces no errors", func(t *testing.T) {
		req := validReq()
		errs := ValidateCreateEmployee(&req)
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got %v", errs)
		}
	})

	t.Run("Missing name", func(t *testing.T) {
		req := validReq()
		req.Name = ""
		errs := ValidateCreateEmployee(&req)
		if _, ok := errs["name"]; !ok {
			t.Error("expected error for missing name")
		}
	})

	t.Run("Invalid email format", func(t *testing.T) {
		req := validReq()
		req.Email = "invalid-email-address"
		errs := ValidateCreateEmployee(&req)
		if _, ok := errs["email"]; !ok {
			t.Error("expected error for invalid email format")
		}
	})

	t.Run("Invalid status", func(t *testing.T) {
		req := validReq()
		req.Status = "Suspended"
		errs := ValidateCreateEmployee(&req)
		if _, ok := errs["status"]; !ok {
			t.Error("expected error for invalid status")
		}
	})

	t.Run("Invalid joining date format", func(t *testing.T) {
		req := validReq()
		req.JoiningDate = "15-01-2024" // Non YYYY-MM-DD
		errs := ValidateCreateEmployee(&req)
		if _, ok := errs["joining_date"]; !ok {
			t.Error("expected error for non YYYY-MM-DD joining date")
		}
	})

	t.Run("Valid Inactive status", func(t *testing.T) {
		req := validReq()
		req.Status = "Inactive"
		errs := ValidateCreateEmployee(&req)
		if len(errs) != 0 {
			t.Errorf("expected Inactive status to be valid, got %v", errs)
		}
	})
}
