package validation

import (
	"order-service/internal/db"

	"github.com/go-playground/validator"
)

var validate = validator.New()

func ValidateOrder(order *db.Order) error {
	return validate.Struct(order)
}
