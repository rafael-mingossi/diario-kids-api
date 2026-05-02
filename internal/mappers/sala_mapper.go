package mappers

// Mapper traduz ids no create; objetos relacionados vêm por preload em leituras

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
)

// CriarSalaInputToModel converte um DTO de entrada em um Model pronto para persistência.
func CriarSalaInputToModel(input dto.CriarSalaInput) models.Sala {
	return models.Sala{
		Nome:        input.Nome,
		Numero:      input.Numero,
		ProfessorID: input.ProfessorID,
	}
}

// ModelToSalaResponse converte um Model de banco em um DTO de resposta seguro.
func ModelToSalaResponse(s models.Sala) dto.SalaResponse {
	return dto.SalaResponse{
		ID:          s.ID, // ID gerado pelo banco apos o INSERT
		Nome:        s.Nome,
		Numero:      s.Numero,
		ProfessorID: s.ProfessorID,
	}
}
