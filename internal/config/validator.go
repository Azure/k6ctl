package config

import "github.com/go-playground/validator/v10"

var defaultValidator = validator.New(validator.WithRequiredStructEnabled())
