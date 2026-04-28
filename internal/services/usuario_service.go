package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/mappers"
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

	// 2. Responsabilidade do Service: criptografia da senha.
	// Isso fica aqui e NÃO no mapper — o mapper é burro por design.
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar senha: %w", err)
	}

	// 3. Delegação ao Mapper: transforma o DTO de entrada + senha hash em um Model de banco.
	// O Service não sabe mais quais campos o model tem — isso é responsabilidade do mapper.
	novoUsuario := mappers.CriarInputToModel(input, string(senhaHash))

	// 4. Delegação ao Repository: persiste o model no banco.
	if err := s.repo.Criar(&novoUsuario); err != nil {
		// Anti-TOCTOU: mesmo que dois requests passem pela checagem de email
		// simultaneamente, o unique index do banco barra o segundo INSERT.
		// O repository traduz o erro do Postgres (código 23505) para ErrEmailDuplicadoDB,
		// e aqui o service traduz para ErrEmailEmUso que o handler já conhece.
		if errors.Is(err, repository.ErrEmailDuplicadoDB) {
			return nil, ErrEmailEmUso
		}
		return nil, fmt.Errorf("erro ao inserir no banco: %w", err)
	}

	// 5. Delegação ao Mapper: transforma o Model salvo (já com ID do banco) em DTO de saída.
	// SenhaHash é omitido pelo mapper — o cliente nunca a recebe.
	resposta := mappers.ModelToUsuarioResponse(novoUsuario)

	return &resposta, nil
}
