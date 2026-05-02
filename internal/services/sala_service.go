package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/mappers"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

type SalaService interface {
	CriarSala(input dto.CriarSalaInput) (dto.SalaResponse, error)
}

type salaService struct {
	repo repository.SalaRepository
}

func NewSalaService(repo repository.SalaRepository) SalaService {
	return &salaService{repo: repo}
}

// erro customizado para o Handler saber quando é conflito
var ErrSalaDuplicada = errors.New("esta sala já existe")

func (s *salaService) CriarSala(input dto.CriarSalaInput) (dto.SalaResponse, error) {
	// 1. Verificar se a sala já existe (por nome e número)
	salaExistente, err := s.repo.BuscarPorNomeENumero(input.Nome, input.Numero)
	if err != nil {
		// Embrulhamos o erro com contexto para facilitar depuração nos logs.
		// %w permite que errors.Is() continue funcionando na cadeia de chamadas.
		return dto.SalaResponse{}, fmt.Errorf("erro ao verificar sala existente: %w", err)
	}
	if salaExistente != nil {
		return dto.SalaResponse{}, ErrSalaDuplicada
	}

	// 2. Converter o DTO de entrada para um Model de banco
	novaSala := mappers.CriarSalaInputToModel(input)

	// 3. Persistir a nova sala no banco
	if err := s.repo.Criar(&novaSala); err != nil {
		if errors.Is(err, repository.ErrSalaDuplicadaDB) {
			return dto.SalaResponse{}, ErrSalaDuplicada
		}
		return dto.SalaResponse{}, fmt.Errorf("erro ao inserir no banco: %w", err)
	}

	// 4. Converter o Model de banco para um DTO de resposta
	resposta := mappers.ModelToSalaResponse(novaSala)

	return resposta, nil
}
