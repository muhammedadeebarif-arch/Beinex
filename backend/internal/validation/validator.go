package validation

import (
	"net/mail"
	"regexp"
	"strings"
	"time"

	"ems-backend/internal/model"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// NormalizeEmail trims and converts email to lower-case.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateCreateEmployee validates a CreateEmployeeRequest and returns a map of field errors if invalid.
func ValidateCreateEmployee(req *model.CreateEmployeeRequest) map[string]string {
	fields := make(map[string]string)

	req.Name = strings.TrimSpace(req.Name)
	req.Email = NormalizeEmail(req.Email)
	req.Department = strings.TrimSpace(req.Department)
	req.Designation = strings.TrimSpace(req.Designation)
	req.JoiningDate = strings.TrimSpace(req.JoiningDate)
	req.Status = strings.TrimSpace(req.Status)

	if req.Name == "" {
		fields["name"] = "Employee name is required."
	} else if len(req.Name) > 100 {
		fields["name"] = "Employee name must not exceed 100 characters."
	}

	if req.Email == "" {
		fields["email"] = "Email address is required."
	} else if len(req.Email) > 150 {
		fields["email"] = "Email address must not exceed 150 characters."
	} else if !emailRegex.MatchString(req.Email) {
		fields["email"] = "Please enter a valid email address."
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		fields["email"] = "Please enter a valid email address."
	}

	if req.Department == "" {
		fields["department"] = "Department is required."
	} else if len(req.Department) > 100 {
		fields["department"] = "Department must not exceed 100 characters."
	}

	if req.Designation == "" {
		fields["designation"] = "Job designation is required."
	} else if len(req.Designation) > 100 {
		fields["designation"] = "Job designation must not exceed 100 characters."
	}

	if req.JoiningDate == "" {
		fields["joining_date"] = "Joining date is required."
	} else {
		parsedDate, err := time.Parse("2006-01-02", req.JoiningDate)
		if err != nil {
			fields["joining_date"] = "Joining date must be in YYYY-MM-DD format."
		} else if parsedDate.Year() < 1970 || parsedDate.Year() > 2100 {
			fields["joining_date"] = "Joining date is outside the acceptable range."
		}
	}

	if req.Status == "" {
		fields["status"] = "Status is required."
	} else if req.Status != "Active" && req.Status != "Inactive" {
		fields["status"] = "Status must be either 'Active' or 'Inactive'."
	}

	if len(fields) == 0 {
		return nil
	}
	return fields
}

// ValidateUpdateEmployee validates an UpdateEmployeeRequest.
func ValidateUpdateEmployee(req *model.UpdateEmployeeRequest) map[string]string {
	createReq := model.CreateEmployeeRequest{
		Name:        req.Name,
		Email:       req.Email,
		Department:  req.Department,
		Designation: req.Designation,
		JoiningDate: req.JoiningDate,
		Status:      req.Status,
	}
	errs := ValidateCreateEmployee(&createReq)
	req.Name = createReq.Name
	req.Email = createReq.Email
	req.Department = createReq.Department
	req.Designation = createReq.Designation
	req.JoiningDate = createReq.JoiningDate
	req.Status = createReq.Status
	return errs
}
