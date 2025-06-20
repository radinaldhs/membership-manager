package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	idTranslation "github.com/go-playground/validator/v10/translations/id"
)

func NewValidator() (*validator.Validate, ut.Translator, error) {
	// We will be using Bahasa Indonesia for our client-facing errors
	idLang := id.New()
	uniTrans := ut.New(idLang, idLang)
	translator, _ := uniTrans.GetTranslator("id")

	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return field.Name
		}
		return name
	})

	err := idTranslation.RegisterDefaultTranslations(validate, translator)
	if err != nil {
		return nil, nil, err
	}

	registerCustomValidations(validate)
	if err := registerCustomTranslations(validate, translator); err != nil {
		return nil, nil, err
	}

	return validate, translator, nil
}
