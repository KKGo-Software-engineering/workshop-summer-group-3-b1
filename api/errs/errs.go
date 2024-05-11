package errs

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Messages []string `json:"messages"`
}

var (
	ErrInternalDatabaseError = errors.New("internal database error")
)

const (
	required = "field %s is required"
	oneof    = "the value of %s must be one of %s"
	gt       = "the value of %s must be greater than %s"
	gte      = "the value of %s must be greater than or equal %s"
	lte      = "the value of %s must be less than or equal %s"
	ltefield = "the value of %s value must be lower than or equal value of field %s"
	unknown  = "unknown error"
)

func getErrMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf(required, fe.Field())
	case "oneof":
		return fmt.Sprintf(oneof, fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf(gt, fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf(gte, fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf(lte, fe.Field(), fe.Param())
	case "ltefield":
		return fmt.Sprintf(ltefield, fe.Field(), fe.Param())
	}

	return unknown
}

func ParseError(err error) ErrorResponse {
	var messages []string

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			messages = append(messages, getErrMsg(fe))
		}
		return ErrorResponse{Messages: messages}
	}

	messages = append(messages, err.Error())
	return ErrorResponse{Messages: messages}
}
