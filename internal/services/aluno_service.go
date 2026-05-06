package services

import (
	"errors"
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

// Erros de domínio esperados: o handler usa errors.Is para transformar isso em 422.
var ErrDataNascimentoInvalida = errors.New("data_nascimento inválida")
var ErrDataNascimentoFutura = errors.New("data_nascimento não pode ser futura")

// service recebe o DTO de input, processa as regras de negocio e devolve o DTO de output
func (s *alunoService) CriarAluno(input dto.CriarAlunoInput) (*dto.AlunoResponse, error) {
	//TODO verificar se aluno ja existe

	// O DTO traz a data como string porque esse formato é mais amigável para o frontend.
	// Aqui convertemos para time.Time, que é o tipo correto para persistir no model.
	// Parse vem ANTES da validação de data futura, porque enquanto a data ainda é string
	// nós não temos um valor temporal real para comparar com "hoje".
	dataNascimento, err := time.ParseInLocation("02/01/2006", input.DataNascimento, time.Local)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDataNascimentoInvalida, err)
	}

	// Como nascimento é uma data sem horário, normalizamos "hoje" para 00:00.
	// Isso evita bugs de fuso/horário ao comparar apenas a parte da data.
	agora := time.Now().In(time.Local)
	hoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, agora.Location())
	if dataNascimento.After(hoje) {
		return nil, ErrDataNascimentoFutura
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
