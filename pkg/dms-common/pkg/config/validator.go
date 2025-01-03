package config

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

var DbNameFormatPattern = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")

func validateDbNameFormat(field validator.FieldLevel) bool {
	return DbNameFormatPattern.MatchString(field.Field().String())
}

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("dbNameFormat", validateDbNameFormat)
}

func Validate(i interface{}) error {
	if validate == nil {
		return nil
	}

	err := validate.Struct(i)
	if err == nil {
		return nil
	}

	return err
}
