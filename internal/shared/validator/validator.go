package validator

import (
	transmodels "github.com/SlayerSv/payments/internal/trans/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func NewValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("account_type", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return transmodels.GetAccountType(val) != transmodels.AccountUnspecified
	})
	validate.RegisterValidation("operation_type", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return transmodels.GetOperationType(val) != transmodels.OperationUnspecified
	})
	validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		id, _ := uuid.Parse(val)
		return id != uuid.Nil
	})
	return validate
}
