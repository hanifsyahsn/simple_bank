package util

import (
	"github.com/go-playground/validator/v10"
)

func ValidatorError(err validator.ValidationErrors) map[string]string {
	out := map[string]string{}

	for _, e := range err {
		msg := ""
		switch e.Tag() {
		case "required":
			msg = "is required"
		case "alphanum":
			msg = "must be alphanumeric"
		case "min":
			msg = "must be at least " + e.Param() + " characters"
		case "email":
			msg = "must be a valid email"
		default:
			msg = "invalid value"
		}
		out[e.Field()] = e.Field() + " " + msg
	}

	return out
}
