package services

import (
	"errors"
	"regexp"
	"strings"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"go.uber.org/zap"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// Validator validates domain objects
type Validator struct {
	logger *zap.Logger
}

// NewValidator creates a new Validator
func NewValidator(logger *zap.Logger) *Validator {
	return &Validator{logger: logger}
}

// ValidateProduct validates a product create request
func (v *Validator) ValidateProduct(req *models.CreateProductRequest) *ValidationResult {
	result := &ValidationResult{IsValid: true, Errors: make([]ValidationError, 0)}

	if strings.TrimSpace(req.SKU) == "" {
		result.addError("sku", "SKU is required")
	} else if len(req.SKU) > 64 {
		result.addError("sku", "SKU must be 64 characters or less")
	}

	if strings.TrimSpace(req.Name) == "" {
		result.addError("name", "Name is required")
	} else if len(req.Name) > 255 {
		result.addError("name", "Name must be 255 characters or less")
	}

	if req.Slug != "" && !isValidSlug(req.Slug) {
		result.addError("slug", "Slug must contain only lowercase letters, numbers, and hyphens")
	}

	if req.Weight < 0 {
		result.addError("weight", "Weight cannot be negative")
	}

	// Validate variants
	for i, variant := range req.Variants {
		if strings.TrimSpace(variant.SKU) == "" {
			result.addError("variants", "Variant SKU is required")
		}
		if variant.Price < 0 {
			result.addErrorf("variants[%d].price", "Price cannot be negative", i)
		}
	}

	return result
}

// ValidateCategory validates a category create request
func (v *Validator) ValidateCategory(req *models.CreateCategoryRequest) *ValidationResult {
	result := &ValidationResult{IsValid: true, Errors: make([]ValidationError, 0)}

	if strings.TrimSpace(req.Name) == "" {
		result.addError("name", "Name is required")
	} else if len(req.Name) > 255 {
		result.addError("name", "Name must be 255 characters or less")
	}

	if req.Slug != "" && !isValidSlug(req.Slug) {
		result.addError("slug", "Slug must contain only lowercase letters, numbers, and hyphens")
	}

	return result
}

// ValidateBrand validates a brand create request
func (v *Validator) ValidateBrand(req *models.CreateBrandRequest) *ValidationResult {
	result := &ValidationResult{IsValid: true, Errors: make([]ValidationError, 0)}

	if strings.TrimSpace(req.Name) == "" {
		result.addError("name", "Name is required")
	} else if len(req.Name) > 255 {
		result.addError("name", "Name must be 255 characters or less")
	}

	if req.WebsiteURL != "" && !isValidURL(req.WebsiteURL) {
		result.addError("website_url", "Invalid URL format")
	}

	return result
}

func (r *ValidationResult) addError(field, message string) {
	r.IsValid = false
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
}

func (r *ValidationResult) addErrorf(field, message string, args ...interface{}) {
	r.addError(field, message)
}

// ToError converts validation result to an error
func (r *ValidationResult) ToError() error {
	if r.IsValid {
		return nil
	}
	messages := make([]string, len(r.Errors))
	for i, e := range r.Errors {
		messages[i] = e.Field + ": " + e.Message
	}
	return errors.New(strings.Join(messages, "; "))
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var urlRegex = regexp.MustCompile(`^https?://`)

func isValidSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}

func isValidURL(url string) bool {
	return urlRegex.MatchString(url)
}

