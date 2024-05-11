package utils

import (
	"github.com/labstack/echo/v4"
)

type KeyValuePairs map[string]string

func SetParams(c echo.Context, data KeyValuePairs) {
	for k, v := range data {
		c.SetParamNames(k)
		c.SetParamValues(v)
	}
}
