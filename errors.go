package main

import "fmt"

// Custom error types for better error handling
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

type AuthorizationError struct {
	Message string
}

func (e AuthorizationError) Error() string {
	return fmt.Sprintf("authorization error: %s", e.Message)
}

type FileProcessingError struct {
	Operation string
	Message   string
}

func (e FileProcessingError) Error() string {
	return fmt.Sprintf("file processing error during '%s': %s", e.Operation, e.Message)
}

type S3Error struct {
	Operation string
	Message   string
}

func (e S3Error) Error() string {
	return fmt.Sprintf("S3 error during '%s': %s", e.Operation, e.Message)
}

// Helper functions to create errors
func NewValidationError(field, message string) ValidationError {
	return ValidationError{Field: field, Message: message}
}

func NewAuthorizationError(message string) AuthorizationError {
	return AuthorizationError{Message: message}
}

func NewFileProcessingError(operation, message string) FileProcessingError {
	return FileProcessingError{Operation: operation, Message: message}
}

func NewS3Error(operation, message string) S3Error {
	return S3Error{Operation: operation, Message: message}
}
