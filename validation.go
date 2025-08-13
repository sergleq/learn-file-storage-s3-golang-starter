package main

import (
	"mime"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
)

// ValidationResult represents the result of a validation
type ValidationResult struct {
	IsValid bool
	Error   error
}

// ValidateVideoID validates if the video ID is a valid UUID
func ValidateVideoID(videoIDString string) ValidationResult {
	if videoIDString == "" {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("videoID", "video ID is required"),
		}
	}

	_, err := uuid.Parse(videoIDString)
	if err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("videoID", "invalid video ID format"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidateVideoFile validates if the uploaded file is a valid video
func ValidateVideoFile(header *multipart.FileHeader) ValidationResult {
	if header == nil {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("video", "video file is required"),
		}
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("video", "invalid content type"),
		}
	}

	if mediaType != VideoMP4Type {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("video", "only MP4 videos are supported"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidateThumbnailFile validates if the uploaded file is a valid thumbnail
func ValidateThumbnailFile(header *multipart.FileHeader) ValidationResult {
	if header == nil {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("thumbnail", "thumbnail file is required"),
		}
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("thumbnail", "invalid content type"),
		}
	}

	if mediaType != ImageJPEGType && mediaType != ImagePNGType {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("thumbnail", "only JPEG and PNG images are supported"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidateVideoURL validates if the video URL has the correct format
func ValidateVideoURL(videoURL string) ValidationResult {
	if videoURL == "" {
		return ValidationResult{IsValid: true, Error: nil} // Empty URL is valid (optional field)
	}

	parts := strings.Split(videoURL, ",")
	if len(parts) != 2 {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("videoURL", "invalid video URL format: expected 'bucket,key'"),
		}
	}

	if parts[0] == "" || parts[1] == "" {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("videoURL", "bucket and key cannot be empty"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidateEmail validates if the email has a valid format
func ValidateEmail(email string) ValidationResult {
	if email == "" {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("email", "email is required"),
		}
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("email", "invalid email format"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidatePassword validates if the password meets requirements
func ValidatePassword(password string) ValidationResult {
	if password == "" {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("password", "password is required"),
		}
	}

	if len(password) < 6 {
		return ValidationResult{
			IsValid: false,
			Error:   NewValidationError("password", "password must be at least 6 characters long"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}
