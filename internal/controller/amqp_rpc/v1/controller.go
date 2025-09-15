package v1

import (
	"github.com/OpenBankingVN/BRIDGE-API/internal/usecase"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	t usecase.Translation
	l logger.Interface
	v *validator.Validate
}
