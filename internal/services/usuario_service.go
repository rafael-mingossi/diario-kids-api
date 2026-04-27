package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// NOVIDADE: O Service agora também tem uma Interface Pública!
type UsuarioService interface {
	CriarUsuario(input dto.CriarUsuarioInput) (*dto.UsuarioResponse, error)
}

// A implementação virou privada (letra minúscula)
// Service is an interface for the same reason — the handler depends on it
type usuarioService struct {
	repo repository.UsuarioRepository
}

// O construtor devolve a Interface
func NewUsuarioService(repo repository.UsuarioRepository) UsuarioService {
	return &usuarioService{repo: repo}
}

// NOVIDADE: Criamos um erro customizado para o Handler saber quando é conflito
var ErrEmailEmUso = errors.New("este email já está cadastrado")

// O Service recebe o DTO de Input, processa as regras de negócio, e devolve o DTO de Output
func (s *usuarioService) CriarUsuario(input dto.CriarUsuarioInput) (*dto.UsuarioResponse, error) {

	// 1. Verificamos se o email já existe (Feedback 2)
	usuarioExistente, err := s.repo.BuscarPorEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar email: %w", err)
	}
	if usuarioExistente != nil {
		return nil, ErrEmailEmUso // Devolvemos nosso erro específico!
	}

	senhaHash, err := bcrypt.GenerateFromPassword([]byte(input.Senha), 10)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar senha: %w", err)
	}

	// Transformação: DTO (Entrada) -> Model (Banco)
	novoUsuario := models.Usuario{
		Nome:      input.Nome,
		Email:     input.Email,
		SenhaHash: string(senhaHash),
		Role:      input.Role,
	}

	// Delegação: Pede ao repositório para salvar o novo usuário
	if err := s.repo.Criar(&novoUsuario); err != nil {
		// Regra 2: Tratamento de Conflito.
		// Avalia o erro técnico do Repositório e o traduz para uma regra de negócio.
		// A nossa trava anti-TOCTOU entra aqui!
		if errors.Is(err, repository.ErrEmailDuplicadoDB) {
			return nil, ErrEmailEmUso // Transforma no erro 409 que o Handler já entende
		}
		// Se for outro erro (ex: banco offline), repassa.
		return nil, fmt.Errorf("erro ao inserir no banco: %w", err)
	}

	// Transformação: Model (Banco) -> DTO (Saída segura)
	resposta := dto.UsuarioResponse{
		ID:    novoUsuario.ID,
		Nome:  novoUsuario.Nome,
		Email: novoUsuario.Email,
		Role:  novoUsuario.Role,
	}

	return &resposta, nil
}
