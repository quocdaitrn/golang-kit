package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	stdvalidator "github.com/go-playground/validator/v10"
	entrans "github.com/go-playground/validator/v10/translations/en"

	kiterrors "github.com/quocdaitrn/golang-kit/errors"
)

// Validator provides available methods to validate struct.
type Validator interface {
	Validate(i interface{}) error
}

// validator represents a validator for request data.
type validator struct {
	Validator  *stdvalidator.Validate
	Translator ut.Translator
}

// New creates and returns a new instance of Validator.
func New() (Validator, error) {
	v := stdvalidator.New()
	// Get name for validator by priorities: header > param > query > json > form.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("field"), ",", 2)[0]
		if name != "" && name != "-" {
			return name
		}

		if name == "-" {
			return ""
		}

		return name
	})
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)

	trans, _ := uni.GetTranslator("en")

	if err := entrans.RegisterDefaultTranslations(v, trans); err != nil {
		return nil, err
	}

	return &validator{Validator: v, Translator: trans}, nil
}

// Validate validate input struct.
func (v *validator) Validate(i interface{}) error {
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

		return kiterrors.WithStack(kiterrors.ErrInvalidRequest.WithDetails(details))
	}

	return nil
}
