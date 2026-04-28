package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Interface publica pois o Handler de autenticação depende dela
type AuthService interface {
	// Recebe o DTO de login e devolve o DTO de resposta ou um erro
	// Estamos indicando que a resposta LoginResponse deve ser um ponteiro, para evitar
	// cópias desnecessárias e também para poder devolver nil em caso de erro.
	Login(input dto.LoginInput) (*dto.LoginResponse, error)
}

// A implementação do AuthService é privada, pois só o Handler de autenticação precisa dela
type authService struct {
	repo repository.UsuarioRepository
}

// O construtor do AuthService recebe a dependência do UsuarioRepository e devolve a interface
func NewAuthService(repo repository.UsuarioRepository) AuthService {
	return &authService{repo: repo}
}

// Implementação do método Login
func (s *authService) Login(input dto.LoginInput) (*dto.LoginResponse, error) {
	// 1. Buscar no banco
	usuario, err := s.repo.BuscarPorEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("erro interno ao buscar usuario: %w", err)
	}

	// 2. Se o usuário for nil (não existe), barra na hora!
	if usuario == nil {
		return nil, errors.New("credenciais inválidas") // Erro vago de propósito
	}

	// 3. Comparar a senha
	err = bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(input.Senha))
	if err != nil {
		return nil, errors.New("credenciais inválidas") // Mesmo erro vago
	}

	// TODO: Gerar JWT

	resposta := dto.LoginResponse{
		Token: "token-fake-para-teste", // Substituir pelo token real depois
		Email: usuario.Email,
	}

	return &resposta, nil
}
