package mappers

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
)

func CriarEscolaInputToModel(input dto.CriarEscolaInput) models.Escola {
	return models.Escola{
		ClienteID:    input.ClienteID,
		CNPJ:         input.CNPJ,
		RazaoSocial:  input.RazaoSocial,
		NomeFantasia: input.NomeFantasia,
		IsMatriz:     input.IsMatriz,
		MatrizID:     input.MatrizID,
	}
}

func ModelToEscolaResponse(e models.Escola) dto.EscolaResponse {
	return dto.EscolaResponse{
		ID:           e.ID,
		ClienteID:    e.ClienteID,
		CNPJ:         e.CNPJ,
		RazaoSocial:  e.RazaoSocial,
		NomeFantasia: e.NomeFantasia,
		IsMatriz:     e.IsMatriz,
		MatrizID:     e.MatrizID,
	}
}
