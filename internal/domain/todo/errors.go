package domain

import "errors"

// Domain errors represent business rule violations and validation failures.
// These are expected errors that are part of the domain model.

var (
	// Validation errors
	ErrInvalidTitle    = errors.New("title must be between 1 and 200 characters")
	ErrInvalidDueDate  = errors.New("due date must be in the future")
	ErrInvalidPriority = errors.New("invalid priority value")
	ErrInvalidStatus   = errors.New("invalid status value")
	ErrInvalidID       = errors.New("invalid todo ID")

	// Business rule errors
	ErrCannotCompleteCancelled = errors.New("cannot complete a cancelled task")
	ErrCannotModifyCompleted   = errors.New("cannot modify a completed task")
	ErrTodoNotFound            = errors.New("todo not found")
	ErrTodoAlreadyExists       = errors.New("todo already exists")

	// State transition errors
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// DomainError interface for type checking domain errors
type DomainError interface {
	error
	IsDomainError()
}

// ValidationError represents a validation failure with context
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func (e ValidationError) IsDomainError() {}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// BusinessRuleError represents a business rule violation
type BusinessRuleError struct {
	Rule    string
	Message string
}

func (e BusinessRuleError) Error() string {
	return "business rule violated (" + e.Rule + "): " + e.Message
}

func (e BusinessRuleError) IsDomainError() {}

// NewBusinessRuleError creates a new business rule error
func NewBusinessRuleError(rule, message string) BusinessRuleError {
	return BusinessRuleError{
		Rule:    rule,
		Message: message,
	}
}
