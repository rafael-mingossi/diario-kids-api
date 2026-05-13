package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/mappers"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

type ClienteService interface {
	CriarCliente(input dto.CriarClienteInput) (dto.ClienteResponse, error)
}

type clienteService struct {
	repo repository.ClienteRepository
}

func NewClienteService(repo repository.ClienteRepository) ClienteService {
	return &clienteService{repo: repo}
}

var ErrClienteDuplicado = errors.New("cliente já existe")
var ErrClienteNaoEncontrado = errors.New("cliente não encontrado")

func (s *clienteService) CriarCliente(input dto.CriarClienteInput) (dto.ClienteResponse, error) {
	clienteExistente, err := s.repo.BuscarPorDocumento(input.Documento)
	if err != nil {
		return dto.ClienteResponse{}, fmt.Errorf("erro ao verificar documento existente: %w", err)
	}
	if clienteExistente != nil {
		return dto.ClienteResponse{}, ErrClienteDuplicado
	}

	novoCliente := mappers.CriarClienteInputToModel(input)
	if err := s.repo.Criar(&novoCliente); err != nil {
		if errors.Is(err, repository.ErrClienteDuplicadoDB) {
			return dto.ClienteResponse{}, ErrClienteDuplicado
		}
		return dto.ClienteResponse{}, fmt.Errorf("erro ao criar cliente: %w", err)
	}

	return mappers.ModelToClienteResponse(novoCliente), nil
}
