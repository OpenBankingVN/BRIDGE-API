package v1

import (
	v1 "github.com/OpenBankingVN/BRIDGE-API/docs/proto/v1"
	"github.com/OpenBankingVN/BRIDGE-API/internal/usecase"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	v1.TranslationServer

	t usecase.Translation
	l logger.Interface
	v *validator.Validate
}
