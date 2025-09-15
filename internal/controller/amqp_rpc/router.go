package v1

import (
	v1 "github.com/OpenBankingVN/BRIDGE-API/internal/controller/amqp_rpc/v1"
	"github.com/OpenBankingVN/BRIDGE-API/internal/usecase"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/logger"
	"github.com/OpenBankingVN/BRIDGE-API/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(t usecase.Translation, l logger.Interface) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)

	{
		v1.NewTranslationRoutes(routes, t, l)
	}

	return routes
}
