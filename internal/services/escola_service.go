package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/mappers"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

type EscolaService interface {
	CriarEscola(input dto.CriarEscolaInput) (dto.EscolaResponse, error)
}

type escolaService struct {
	repo repository.EscolaRepository
}

func NewEscolaService(repo repository.EscolaRepository) EscolaService {
	return &escolaService{repo: repo}
}

// Sentinels compartilhadas neste pacote.
// `ErrEscolaNaoEncontrada` também será reutilizada por Sala e Aluno para classificar 422.
var ErrEscolaDuplicada = errors.New("escola já existe")
var ErrEscolaNaoEncontrada = errors.New("escola não encontrada")
var ErrMatrizObrigatoria = errors.New("filial precisa informar matriz_id")
var ErrMatrizInvalida = errors.New("matriz informada é inválida")
var ErrMatrizNaoPodeTerMatriz = errors.New("escola matriz não pode informar matriz_id")

func (s *escolaService) CriarEscola(input dto.CriarEscolaInput) (dto.EscolaResponse, error) {
	// CNPJ é o identificador de negócio mais forte da escola/unidade.
	escolaExistente, err := s.repo.BuscarPorCNPJ(input.CNPJ)
	if err != nil {
		return dto.EscolaResponse{}, fmt.Errorf("erro ao verificar cnpj existente: %w", err)
	}
	if escolaExistente != nil {
		return dto.EscolaResponse{}, ErrEscolaDuplicada
	}

	// Regra de domínio da hierarquia:
	// - matriz: não aponta para outra matriz
	// - filial: obrigatoriamente aponta para uma matriz existente
	if input.IsMatriz {
		if input.MatrizID != nil {
			return dto.EscolaResponse{}, ErrMatrizNaoPodeTerMatriz
		}
	} else {
		if input.MatrizID == nil {
			return dto.EscolaResponse{}, ErrMatrizObrigatoria
		}

		matriz, err := s.repo.BuscarPorID(*input.MatrizID)
		if err != nil {
			return dto.EscolaResponse{}, fmt.Errorf("erro ao verificar matriz: %w", err)
		}
		if matriz == nil || !matriz.IsMatriz {
			return dto.EscolaResponse{}, ErrMatrizInvalida
		}
	}

	novaEscola := mappers.CriarEscolaInputToModel(input)
	if err := s.repo.Criar(&novaEscola); err != nil {
		if errors.Is(err, repository.ErrEscolaDuplicadaDB) {
			return dto.EscolaResponse{}, ErrEscolaDuplicada
		}
		return dto.EscolaResponse{}, fmt.Errorf("erro ao criar escola: %w", err)
	}

	return mappers.ModelToEscolaResponse(novaEscola), nil
}