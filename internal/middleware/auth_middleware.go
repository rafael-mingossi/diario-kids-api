// Package middleware contém middlewares HTTP reutilizáveis da API.
// Um middleware é uma função que intercepta a requisição ANTES de chegar no handler
// e pode permitir, rejeitar ou enriquecer a requisição.
//
// Analogia JS: é o equivalente a um app.use() ou router.use() do Express.js —
// uma função que fica "no meio do caminho" entre o cliente e o controller.
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey é um tipo próprio para chaves de contexto.
//
// Por que não usar uma string simples?
// Em Go, o contexto de requisição (r.Context()) é um mapa de interface{} -> interface{}.
// Se usarmos uma string nua como chave (ex: "usuario"), qualquer outro pacote que também
// use a chave "usuario" irá sobrescrever nosso valor sem erros em tempo de compilação.
// Com um tipo próprio, a colisão é impossível — dois pacotes nunca terão o mesmo tipo.
type contextKey string

// ContextKeyUsuario é a chave pública exportada para que handlers possam extrair
// os dados do usuário autenticado do contexto da requisição.
//
// Uso em um handler protegido:
//
//	usuario := r.Context().Value(middleware.ContextKeyUsuario).(middleware.UsuarioAutenticado)
const ContextKeyUsuario contextKey = "usuario"

// UsuarioAutenticado representa os dados do usuário que ficam disponíveis
// para qualquer handler depois que o middleware valida o token.
//
// Analogia JS: é o objeto que você coloca em `req.user = { id, email, role }` no Express.
type UsuarioAutenticado struct {
	ID    string // ID do banco de dados (string porque foi armazenado no campo Subject do JWT)
	Email string
	Role  string
}

// UsuarioClaims define o formato esperado do payload do JWT durante a validação.
// Deve espelhar exatamente o que o AuthService assinou em gerarJWT().
//
// jwt.RegisteredClaims inclui os campos padrão do protocolo:
// sub, iat, exp — que a biblioteca verifica automaticamente.
type UsuarioClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// Verificar retorna um middleware que valida o JWT de cada requisição.
// Recebe o segredo JWT (lido do ambiente pelo main.go) e devolve uma função de middleware.
//
// Por que receber o secret como parâmetro em vez de ler do os.Getenv aqui?
// Porque injetar a dependência via parâmetro torna o middleware testável de forma isolada —
// você pode passar qualquer segredo nos testes sem precisar setar variáveis de ambiente.
// É o mesmo princípio de injeção de dependência que usamos nos services e repositories.
func Verificar(secret string) func(http.Handler) http.Handler {
	// Retornamos uma função que envolve o próximo handler — este é o padrão
	// de middleware em Go: uma função que recebe o "próximo" e devolve um novo handler.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Passo 1: Lemos o header "Authorization" da requisição HTTP.
			// O cliente deve enviar: Authorization: Bearer eyJhbGci...
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "token de autenticação ausente", http.StatusUnauthorized)
				return
			}

			// Passo 2: Dividimos o header em duas partes: ["Bearer", "<token>"]
			// SplitN com limite 2 evita problemas se o token contiver espaços acidentais
			partes := strings.SplitN(authHeader, " ", 2)
			if len(partes) != 2 || partes[0] != "Bearer" {
				http.Error(w, "formato inválido (esperado: Bearer <token>)", http.StatusUnauthorized)
				return
			}

			tokenString := partes[1]

			// Passo 3: Fazemos o parse e a validação completa do token.
			// A biblioteca verifica automaticamente:
			// - Formato (3 partes separadas por ponto)
			// - Assinatura (usando o secret que fornecemos na função de keyfunc)
			// - Expiração (campo exp do RegisteredClaims)
			token, err := jwt.ParseWithClaims(
				tokenString,
				&UsuarioClaims{}, // O tipo para deserializar o payload
				func(token *jwt.Token) (interface{}, error) {
					// Segurança crítica: verificamos explicitamente o algoritmo de assinatura.
					// Sem esta checagem, um atacante poderia enviar um token assinado com
					// o algoritmo "none" (sem assinatura) e contornar toda a verificação.
					// Sempre validar que é HMAC antes de confiar no token.
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("algoritmo de assinatura inesperado no token")
					}
					// Devolvemos o segredo para a biblioteca verificar a assinatura
					return []byte(secret), nil
				},
			)

			// Se o parse falhou (token expirado, assinatura inválida, malformado, etc.)
			// logamos o motivo internamente mas devolvemos uma mensagem genérica ao cliente.
			if err != nil || !token.Valid {
				slog.Warn("requisição rejeitada: token JWT inválido",
					"ip", r.RemoteAddr,
					"path", r.URL.Path,
					"erro", err,
				)
				http.Error(w, "token inválido ou expirado", http.StatusUnauthorized)
				return
			}

			// Passo 4: Extraímos as claims (o payload do token) com type assertion.
			// O ok verifica se a conversão de tipo foi bem-sucedida.
			claims, ok := token.Claims.(*UsuarioClaims)
			if !ok {
				http.Error(w, "token inválido", http.StatusUnauthorized)
				return
			}

			// Passo 5: Injetamos os dados do usuário no contexto da requisição.
			// context.WithValue cria uma cópia do contexto com o novo par chave-valor.
			// Analogia JS: é o `req.user = { id, email, role }` do Express — a partir
			// daqui, qualquer handler downstream pode acessar o usuário autenticado.
			ctx := context.WithValue(r.Context(), ContextKeyUsuario, UsuarioAutenticado{
				ID:    claims.Subject, // Subject = ID do usuário (definido em gerarJWT)
				Email: claims.Email,
				Role:  claims.Role,
			})

			// Passo 6: Chamamos o próximo handler com o contexto enriquecido.
			// r.WithContext(ctx) cria uma cópia da requisição com o novo contexto.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
