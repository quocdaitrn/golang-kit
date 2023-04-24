package validator

import (
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	stdvalidator "github.com/go-playground/validator/v10"
	entrans "github.com/go-playground/validator/v10/translations/en"
)

// Validator provides available methods to validate struct.
type Validator interface {
	Validate(i interface{}) map[string]string
}

// validator represents a validator for request data.
type validator struct {
	Validator  *stdvalidator.Validate
	Translator ut.Translator
}

// New creates and returns a new instance of Validator.
func New() (Validator, error) {
	v := stdvalidator.New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)

	trans, _ := uni.GetTranslator("en")

	if err := entrans.RegisterDefaultTranslations(v, trans); err != nil {
		return nil, err
	}

	return &validator{Validator: v, Translator: trans}, nil
}

// Validate validate input struct.
func (v *validator) Validate(i interface{}) map[string]string {
	err := v.Validator.Struct(i)
	if err != nil {
		details := map[string]string{}
		ves := err.(stdvalidator.ValidationErrors)
		for _, ve := range ves {
			ns := ve.Namespace()
			idx := strings.IndexByte(ns, byte('.'))
			ns = ns[idx+1:]

			if _, ok := details[ns]; !ok {
				details[ns] = ve.Translate(v.Translator)
			}
		}

		return details
	}

	return nil
}
