package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse представляет единый формат ошибки
type ErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

// SimpleErrorResponse для простых ошибок
type SimpleErrorResponse struct {
	Error string `json:"error"`
}

// FormatValidationErrors форматирует ошибки валидации в единый формат
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := strings.ToLower(fieldError.Field())
			
			// Преобразуем имя поля из CamelCase в snake_case
			switch fieldName {
			case "originalurl":
				fieldName = "original_url"
			case "shortname":
				fieldName = "short_name"
			}

			errors[fieldName] = formatFieldError(fieldError)
		}
	}

	return errors
}

// formatFieldError форматирует конкретную ошибку поля
func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "url":
		return "must be a valid URL"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "alphanum":
		return "must contain only alphanumeric characters"
	default:
		return fmt.Sprintf("validation failed on '%s' tag", fe.Tag())
	}
}

// IsDuplicateKeyError проверяет, является ли ошибка конфликтом уникальности
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL код ошибки 23505 - unique_violation
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
		strings.Contains(err.Error(), "23505")
}

// FormatDuplicateKeyError форматирует ошибку дублирования ключа
func FormatDuplicateKeyError(err error, field string) map[string]string {
	errors := make(map[string]string)
	
	if strings.Contains(err.Error(), "short_name") {
		errors["short_name"] = "short name already in use"
	} else {
		errors[field] = "value already in use"
	}
	
	return errors
}

