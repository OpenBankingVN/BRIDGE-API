package v1

import (
	"github.com/OpenBankingVN/BRIDGE-API/internal/usecase"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/logger"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/rabbitmq/rmq_rpc/server"
	"github.com/go-playground/validator/v10"
)

// NewTranslationRoutes -.
func NewTranslationRoutes(routes map[string]server.CallHandler, t usecase.Translation, l logger.Interface) {
	r := &V1{t: t, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	{
		routes["v1.getHistory"] = r.getHistory()
	}
}
