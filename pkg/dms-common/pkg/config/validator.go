package config

import validator "github.com/go-playground/validator/v10"

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = validator.New()
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
