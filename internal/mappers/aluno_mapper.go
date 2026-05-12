package mappers

import (
	"time"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
)

// CriarAlunoInputToModel converte o DTO de entrada em um Model.
// Repare que o mapper recebe a data já convertida para time.Time.
// O parse da string NÃO acontece aqui; essa decisão fica no service.
func CriarAlunoInputToModel(input dto.CriarAlunoInput, dataNascimento time.Time) models.Aluno {
	escolaID := input.EscolaID
	return models.Aluno{
		EscolaID:       &escolaID,
		Nome:           input.Nome,
		DataNascimento: dataNascimento,
		SalaID:         input.SalaID,
	}
}

// ModelToAlunoResponse converte um Model de banco em um DTO de resposta seguro.
// Aqui fazemos o caminho inverso: time.Time -> string para devolver ao frontend.
func ModelToAlunoResponse(a models.Aluno) dto.AlunoResponse {
	return dto.AlunoResponse{
		ID:             a.ID,
		EscolaID:       a.EscolaID,
		Nome:           a.Nome,
		DataNascimento: a.DataNascimento.Format("02/01/2006"),
		SalaID:         a.SalaID,
		Matricula:      a.Matricula,
	}
}
