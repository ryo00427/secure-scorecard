package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	apperrors "github.com/secure-scorecard/backend/internal/errors"
)

// CustomValidator wraps go-playground/validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom validators here if needed
	// Example: v.RegisterValidation("custom_rule", customRuleFunc)

	return &CustomValidator{validator: v}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Convert validator errors to AppError
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			details := make([]map[string]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				details = append(details, map[string]string{
					"field":   toSnakeCase(e.Field()),
					"tag":     e.Tag(),
					"value":   fmt.Sprintf("%v", e.Value()),
					"message": getErrorMessage(e),
				})
			}
			return apperrors.NewValidationError("Validation failed", details)
		}
		return err
	}
	return nil
}

// BindAndValidate binds request body and validates it
func BindAndValidate(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return apperrors.NewBadRequestError("Invalid request body")
	}
	if err := c.Validate(i); err != nil {
		return err
	}
	return nil
}

// getErrorMessage returns a human-readable error message for validation errors
func getErrorMessage(e validator.FieldError) string {
	field := toSnakeCase(e.Field())
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
