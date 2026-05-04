package services

import (
	"fmt"
	"time"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/mappers"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

// interface publica
type AlunoService interface {
	CriarAluno(input dto.CriarAlunoInput) (*dto.AlunoResponse, error)
}

// interface privada
type alunoService struct {
	repo repository.AlunoRepository
}

// construtor devolve a interfaxe
func NewAlunoService(repo repository.AlunoRepository) AlunoService {
	return &alunoService{repo: repo}
}

// service recebe o DTO de input, processa as regras de negocio e devolve o DTO de output
func (s *alunoService) CriarAluno(input dto.CriarAlunoInput) (*dto.AlunoResponse, error) {
	//TODO verificar se aluno ja existe

	// O DTO traz a data como string porque esse formato é mais amigável para o frontend.
	// Aqui convertemos para time.Time, que é o tipo correto para persistir no model.
	dataNascimento, err := time.Parse("02/01/2006", input.DataNascimento)
	if err != nil {
		return nil, fmt.Errorf("data_nascimento inválida: %w", err)
	}

	// Com a data já convertida, o mapper só monta o model sem regra extra.
	novoAluno := mappers.CriarAlunoInputToModel(input, dataNascimento)

	//persiste model no banco
	if err := s.repo.Criar(&novoAluno); err != nil {
		return nil, fmt.Errorf("erro ao criar aluno: %w", err)
	}

	//converte model para DTO de resposta
	response := mappers.ModelToAlunoResponse(novoAluno)

	return &response, nil
}
