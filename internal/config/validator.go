package config

import "github.com/go-playground/validator/v10"

// DefaultValidator is the default validator used by the config package.
var DefaultValidator = validator.New(validator.WithRequiredStructEnabled())
