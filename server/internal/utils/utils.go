package utils

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func SuccessResponse(data interface{}) fiber.Map {
	return fiber.Map{
		"success": true,
		"data":    data,
	}
}

func ErrorResponse(message string) fiber.Map {
	return fiber.Map{
		"success": false,
		"error":   message,
	}
}

func ValidationErrorResponse(errors []*ErrorResponse) fiber.Map {
	return fiber.Map{
		"success": false,
		"errors":  errors,
	}
}

func ValidateStruct(s interface{}) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructField()
			element.Message = formatValidationError(err)
			errors = append(errors, &element)
		}
	}
	return errors
}

func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Minimum length is " + err.Param()
	case "max":
		return "Maximum length is " + err.Param()
	}
	return err.Tag() + " validation failed"
}

func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}
