package utils

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct checks structural tag validation
func ValidateStruct(s any) error {
	return validate.Struct(s)
}
