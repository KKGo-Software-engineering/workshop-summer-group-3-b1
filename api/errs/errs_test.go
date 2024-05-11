package errs

import (
	ut "github.com/go-playground/universal-translator"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type mockFieldError struct {
	tag   string
	field string
	param string
}

func (m mockFieldError) Namespace() string {
	return ""
}

func (m mockFieldError) StructNamespace() string {
	return ""
}

func (m mockFieldError) StructField() string {
	return ""
}

func (m mockFieldError) Error() string {
	return ""
}

func (m mockFieldError) Tag() string {
	return m.tag
}

func (m mockFieldError) Field() string {
	return m.field
}

func (m mockFieldError) Param() string {
	return m.param
}

func (m mockFieldError) Kind() reflect.Kind {
	return reflect.String
}

func (m mockFieldError) Type() reflect.Type {
	return nil
}

func (m mockFieldError) Value() interface{} {
	return nil
}

func (m mockFieldError) Translate(trans ut.Translator) string {
	return ""
}

func (m mockFieldError) ActualTag() string {
	return m.tag
}

func TestGetErrMsg(t *testing.T) {
	tests := []struct {
		tag      string
		field    string
		param    string
		expected string
	}{
		{"required", "Name", "", "field Name is required"},
		{"oneof", "State", "NY CA TX", "the value of State must be one of NY CA TX"},
		{"gt", "Age", "18", "the value of Age must be greater than 18"},
		{"gte", "Members", "1", "the value of Members must be greater than or equal 1"},
		{"ltefield", "StartYear", "EndYear", "the value of StartYear value must be lower than or equal value of field EndYear"},
		{"lte", "Age", "18", "the value of Age must be less than or equal 18"},
		{"unknown", "Field", "Param", unknown},
	}

	for _, tt := range tests {
		fe := mockFieldError{tag: tt.tag, field: tt.field, param: tt.param}
		msg := getErrMsg(fe)
		assert.Equal(t, tt.expected, msg)
	}
}

func TestParseError(t *testing.T) {
	t.Run("given validation error should return error messages", func(t *testing.T) {
		ve := validator.ValidationErrors{
			mockFieldError{
				tag:   "required",
				field: "Name",
			},
			mockFieldError{
				tag:   "gte",
				field: "Age",
				param: "21",
			},
		}

		resp := ParseError(ve)
		expected := ErrorResponse{
			Messages: []string{
				"field Name is required",
				"the value of Age must be greater than or equal 21",
			},
		}

		assert.Equal(t, expected, resp)
	})

	t.Run("given other error should return error messages", func(t *testing.T) {
		resp := ParseError(ErrInternalDatabaseError)
		expected := ErrorResponse{
			Messages: []string{"internal database error"},
		}

		assert.Equal(t, expected, resp)
	})
}
