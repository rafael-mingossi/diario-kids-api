package services

import (
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// O Service depende da Interface do Repository!
type UsuarioService struct {
	repo repository.UsuarioRepository
}

// O Construtor do Service recebe a Interface do Repository
func NewUsuarioService(repo repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: repo}
}

// O Service recebe o DTO de Input, processa as regras de negócio, e devolve o DTO de Output
func (s *UsuarioService) CriarUsuario(input dto.CriarUsuarioInput) (*dto.UsuarioResponse, error) {

	// 1. Regra de Negócio: Criptografia (Custo 10 está ótimo para começar)
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(input.Senha), 10)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar senha: %w", err)
	}

	// 2. Monta o modelo de banco de dados
	novoUsuario := models.Usuario{
		Nome:      input.Nome,
		Email:     input.Email,
		SenhaHash: string(senhaHash),
		Role:      input.Role,
	}

	// 3. Pede para o Repositório salvar
	if err := s.repo.Criar(&novoUsuario); err != nil {
		return nil, fmt.Errorf("erro no banco de dados (email duplicado?): %w", err)
	}

	// 4. Monta a resposta limpa (sem a senha!)
	resposta := dto.UsuarioResponse{
		ID:    novoUsuario.ID,
		Nome:  novoUsuario.Nome,
		Email: novoUsuario.Email,
		Role:  novoUsuario.Role,
	}

	return &resposta, nil
}
