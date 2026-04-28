package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	// Biblioteca oficial para JWT em Go — equivalente ao 'jsonwebtoken' do Node.js
	"github.com/golang-jwt/jwt/v5"
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Interface publica pois o Handler de autenticação depende dela
type AuthService interface {
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

// ErrCredenciaisInvalidas é o erro exportado (público) para falhas de autenticação
// esperadas: email não encontrado ou senha incorreta.
//
// Por que exportar este erro específico?
// O handler precisa distinguir dois cenários completamente diferentes:
//   - Credenciais erradas → 401 Unauthorized (problema do cliente, mensagem vaga)
//   - Erro interno (DB fora, JWT_SECRET ausente) → 500 Internal Server Error (problema nosso, logar)
//
// Analogia JS: é como ter uma classe customizada de erro `InvalidCredentialsError`
// que você captura separadamente no catch para retornar 401 vs 500.
var ErrCredenciaisInvalidas = errors.New("credenciais inválidas")

// UsuarioClaims define o "recheio" (payload) do nosso JWT.
// Analogia JS: é o objeto { sub, email, role, iat, exp } que você passaria
// para jwt.sign() no Node.js.
//
// jwt.RegisteredClaims já inclui os campos padrão do protocolo JWT:
//   - sub (Subject): quem é o dono do token
//   - iat (IssuedAt): quando foi criado
//   - exp (ExpiresAt): quando expira
type UsuarioClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	// Embedding: ao incorporar RegisteredClaims, herdamos os campos padrão
	// sem precisar redeclará-los — similar ao extends em TypeScript.
	jwt.RegisteredClaims
}

// gerarJWT é uma função interna (letra minúscula = privada ao pacote).
// Ela cria e assina o token com os dados do usuário autenticado.
// Recebe: ID, email e role do usuário.
// Devolve: a string final do token ("header.payload.assinatura") ou um erro.
func gerarJWT(id uint, email, role string) (string, error) {
	// Lemos o segredo do ambiente. NUNCA o embutimos diretamente no código —
	// isso seria o equivalente a commitar uma senha no GitHub.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET não configurado nas variáveis de ambiente")
	}

	// Montamos o payload do token com as claims escolhidas
	claims := UsuarioClaims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			// Subject identifica o dono do token — usamos o ID do banco como string
			Subject: fmt.Sprintf("%d", id),
			// IssuedAt: timestamp de criação (agora)
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// ExpiresAt: o token expira em 24 horas
			// Em produção futura, isso pode vir de uma variável de ambiente (JWT_TTL_MINUTES)
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	// Criamos o token com o algoritmo HMAC-SHA256 (HS256).
	// HS256 usa o mesmo segredo para assinar e verificar — adequado para um servidor único.
	// Em arquiteturas com múltiplos serviços, usaríamos RS256 (chave pública/privada).
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Assinamos o token com o segredo. Este passo gera a 3ª parte da string JWT:
	// "eyJhbGci..." + "." + "eyJzdWIi..." + "." + "<assinatura>"
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("erro ao assinar o token JWT: %w", err)
	}

	return tokenString, nil
}

// Login implementa a regra de negócio de autenticação
func (s *authService) Login(input dto.LoginInput) (*dto.LoginResponse, error) {
	// 1. Buscar o usuário no banco pelo email
	usuario, err := s.repo.BuscarPorEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("erro interno ao buscar usuario: %w", err)
	}

	// 2. Se o usuário não existe, barramos com erro vago de propósito.
	// Nunca dizemos "email não encontrado" — isso ajudaria atacantes a descobrir
	// quais emails estão cadastrados (account enumeration attack).
	if usuario == nil {
		return nil, ErrCredenciaisInvalidas
	}

	// 3. Comparar a senha enviada com o hash armazenado no banco.
	// bcrypt.CompareHashAndPassword é resistente a timing attacks por design.
	err = bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(input.Senha))
	if err != nil {
		// Mesmo erro vago — o cliente não sabe se foi o email ou a senha que estava errada
		return nil, ErrCredenciaisInvalidas
	}

	// 4. Credenciais válidas — geramos o JWT real com os dados do usuário
	tokenString, err := gerarJWT(usuario.ID, usuario.Email, usuario.Role)
	if err != nil {
		// Erro interno (ex: JWT_SECRET ausente). Repassamos para o handler logar.
		return nil, fmt.Errorf("erro ao gerar token: %w", err)
	}

	resposta := dto.LoginResponse{
		Token: tokenString, // Token real assinado e com expiração
		Email: usuario.Email,
	}

	return &resposta, nil
}
