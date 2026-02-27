package validator

import "github.com/go-playground/validator/v10"

type Validator interface {
	Struct(s interface{}) error
}

type validatorImpl struct {
	validate *validator.Validate
}

func New() Validator {
	v := validator.New()
	return &validatorImpl{validate: v}
}

func (v *validatorImpl) Struct(genericStruct interface{}) error {
	return v.validate.Struct(genericStruct)
}
