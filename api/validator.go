package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/projects/go/01_simple_bank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// check currency is supported
		return util.ISSupportedCurrency(currency)
	}

	return false
}
