package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

type AlunoResponsavelService interface {
	Vincular(input dto.VincularAlunoResponsavelInput) (*dto.AlunoResponsavelResponse, error)
}

type alunoResponsavelService struct {
	repo              repository.AlunoResponsavelRepository
	alunoRepo         repository.AlunoRepository
	usuarioRepo       repository.UsuarioRepository
	usuarioEscolaRepo repository.UsuarioEscolaRepository
}

func NewAlunoResponsavelService(repo repository.AlunoResponsavelRepository, alunoRepo repository.AlunoRepository, usuarioRepo repository.UsuarioRepository, usuarioEscolaRepo repository.UsuarioEscolaRepository) AlunoResponsavelService {
	return &alunoResponsavelService{repo: repo, alunoRepo: alunoRepo, usuarioRepo: usuarioRepo, usuarioEscolaRepo: usuarioEscolaRepo}
}

var ErrAlunoResponsavelDuplicado = errors.New("responsável já vinculado a este aluno")
var ErrAlunoNaoEncontrado = errors.New("aluno não encontrado")
var ErrResponsavelNaoEncontrado = errors.New("usuário responsável não encontrado")
var ErrResponsavelNaoPertenceAEscola = errors.New("usuário responsável não pertence à escola do aluno")
var ErrUsuarioNaoEhResponsavel = errors.New("usuário informado não possui papel de responsável na escola")

func (s *alunoResponsavelService) Vincular(input dto.VincularAlunoResponsavelInput) (*dto.AlunoResponsavelResponse, error) {
	aluno, err := s.alunoRepo.BuscarPorID(input.AlunoID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar aluno: %w", err)
	}
	if aluno == nil {
		return nil, ErrAlunoNaoEncontrado
	}

	usuario, err := s.usuarioRepo.BuscarPorID(input.UsuarioID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar usuário responsável: %w", err)
	}
	if usuario == nil {
		return nil, ErrResponsavelNaoEncontrado
	}

	vinculoEscola, err := s.usuarioEscolaRepo.BuscarPorUsuarioEEscola(input.UsuarioID, aluno.EscolaID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar vínculo do responsável com a escola: %w", err)
	}
	if vinculoEscola == nil || !vinculoEscola.Ativo {
		return nil, ErrResponsavelNaoPertenceAEscola
	}
	if vinculoEscola.Role != "responsavel" {
		return nil, ErrUsuarioNaoEhResponsavel
	}

	vinculoExistente, err := s.repo.BuscarPorAlunoEUsuario(input.AlunoID, input.UsuarioID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar vínculo aluno-responsável existente: %w", err)
	}
	if vinculoExistente != nil {
		return nil, ErrAlunoResponsavelDuplicado
	}

	ativo := true
	if input.Ativo != nil {
		ativo = *input.Ativo
	}

	vinculo := models.AlunoResponsavel{
		AlunoID:            input.AlunoID,
		UsuarioID:          input.UsuarioID,
		Parentesco:         input.Parentesco,
		ResponsavelLegal:   input.ResponsavelLegal,
		RecebeNotificacoes: input.RecebeNotificacoes,
		ContatoEmergencia:  input.ContatoEmergencia,
		AutorizadoBusca:    input.AutorizadoBusca,
		Observacao:         input.Observacao,
		Ativo:              ativo,
	}

	if err := s.repo.Criar(&vinculo); err != nil {
		if errors.Is(err, repository.ErrAlunoResponsavelDuplicadoDB) {
			return nil, ErrAlunoResponsavelDuplicado
		}
		return nil, fmt.Errorf("erro ao criar vínculo aluno-responsável: %w", err)
	}

	return &dto.AlunoResponsavelResponse{
		ID:                 vinculo.ID,
		AlunoID:            vinculo.AlunoID,
		UsuarioID:          vinculo.UsuarioID,
		EscolaID:           aluno.EscolaID,
		Parentesco:         vinculo.Parentesco,
		ResponsavelLegal:   vinculo.ResponsavelLegal,
		RecebeNotificacoes: vinculo.RecebeNotificacoes,
		ContatoEmergencia:  vinculo.ContatoEmergencia,
		AutorizadoBusca:    vinculo.AutorizadoBusca,
		Observacao:         vinculo.Observacao,
		Ativo:              vinculo.Ativo,
	}, nil
}
