package validator

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/values"
)

const (
	validationKeyPhoneNum = "phone_num"
)

func registerCustomValidations(validate *validator.Validate) {
	validate.RegisterValidation(validationKeyPhoneNum, validatePhoneNumber)
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		return values.IsPhoneNumberValid(fl.Field().String())
	}

	return true
}
