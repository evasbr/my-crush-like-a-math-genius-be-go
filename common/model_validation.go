// Package common provides cross-cutting utility helper functions
// that can be utilized across all layers of the application.
package common

import (
	"encoding/json"
	"evasbr/mclamg/exception"
	"github.com/go-playground/validator/v10"
)

// Validate checks the fields of a model struct (DTO) based on 'validate' struct tags.
// If any field fails to satisfy its validation rules, this function triggers an
// exception.ValidationError (panic) carrying a JSON list of invalid fields.
//
// Struct Definition Example:
//
//	type CreateProductRequest struct {
//	    Name  string `json:"name" validate:"required"`
//	    Price int64  `json:"price" validate:"required,min=100"`
//	}
//
// Service Layer Usage Example:
//
//	func (s *productService) Create(ctx context.Context, request model.CreateProductRequest) {
//	    common.Validate(request) // Triggers validation panic if request.Name is empty or request.Price < 100
//	    // ... continue business logic ...
//	}
func Validate(modelValidate interface{}) {
	validate := validator.New()
	err := validate.Struct(modelValidate)
	if err != nil {
		var messages []map[string]interface{}
		for _, err := range err.(validator.ValidationErrors) {
			messages = append(messages, map[string]interface{}{
				"field":   err.Field(),
				"message": "this field is " + err.Tag(),
			})
		}

		jsonMessage, errJson := json.Marshal(messages)
		exception.PanicLogging(errJson)

		panic(exception.ValidationError{
			Message: string(jsonMessage),
		})
	}
}
