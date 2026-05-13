package mappers

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
)

func CriarClienteInputToModel(input dto.CriarClienteInput) models.Cliente {
	status := input.Status
	if status == "" {
		status = "ativo"
	}

	plano := input.Plano
	if plano == "" {
		plano = "basico"
	}

	return models.Cliente{
		Nome:            input.Nome,
		Documento:       input.Documento,
		EmailFinanceiro: input.EmailFinanceiro,
		Telefone:        input.Telefone,
		Status:          status,
		Plano:           plano,
	}
}

func ModelToClienteResponse(cliente models.Cliente) dto.ClienteResponse {
	return dto.ClienteResponse{
		ID:              cliente.ID,
		Nome:            cliente.Nome,
		Documento:       cliente.Documento,
		EmailFinanceiro: cliente.EmailFinanceiro,
		Telefone:        cliente.Telefone,
		Status:          cliente.Status,
		Plano:           cliente.Plano,
	}
}
