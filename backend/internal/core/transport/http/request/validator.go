package request

import "github.com/go-playground/validator/v10"

var requestValidator = validator.New()

func Validate(dest any) error {
	return requestValidator.Struct(dest)
}
