package validator

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func registerCustomTranslations(validate *validator.Validate, trans ut.Translator) error {
	if err := phoneNumberValidationTranslate(validate, trans); err != nil {
		return err
	}

	if err := requiredValidationTranslate(validate, trans); err != nil {
		return err
	}

	return nil
}

func phoneNumberValidationTranslate(validate *validator.Validate, trans ut.Translator) error {
	return validate.RegisterTranslation(validationKeyPhoneNum, trans, func(trans ut.Translator) error {
		return trans.Add(validationKeyPhoneNum, "Format nomor telepon tidak valid", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(validationKeyPhoneNum)
		return t
	})
}

func requiredValidationTranslate(validate *validator.Validate, trans ut.Translator) error {
	return validate.RegisterTranslation("required", trans, func(trans ut.Translator) error {
		return trans.Add("required", "Wajib diisi", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required")
		return t
	})
}
