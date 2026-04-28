package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/services"
)

type AuthHandler struct {
	service  services.AuthService
	validate *validator.Validate
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginInput

	// 1. Lê o JSON do corpo da requisição
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON mal formatado", http.StatusBadRequest)
		return
	}

	// 2. Valida os campos usando as tags do DTO (required, email, min=6, etc.)
	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "dados de entrada inválidos", http.StatusUnprocessableEntity)
		return
	}

	// 3. Delega ao Service — ele verifica email, compara bcrypt e gera o JWT
	resposta, err := h.service.Login(input)
	if err != nil {
		// Classificamos o erro usando errors.Is com o sentinel exportado pelo service.
		// Analogia JS: é como o `err instanceof InvalidCredentialsError` no catch.
		//
		// Caso A: credenciais erradas (email não existe ou senha incorreta)
		// → 401 Unauthorized. A mensagem genérica é intencional: nunca revelamos
		//   ao cliente QUAL das duas informações estava errada.
		if errors.Is(err, services.ErrCredenciaisInvalidas) {
			http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
			return
		}

		// Caso B: erro interno inesperado (banco fora, JWT_SECRET ausente, etc.)
		// → Logamos os detalhes técnicos no servidor (para nós depurarmos),
		//   mas devolvemos uma mensagem completamente genérica ao cliente.
		// Nunca exponha err.Error() diretamente — pode vazar detalhes internos.
		slog.Error("erro interno no login",
			"path", r.URL.Path,
			"ip", r.RemoteAddr,
			"detalhe", err, // O detalhe real fica APENAS nos logs do servidor
		)
		http.Error(w, "erro interno no servidor", http.StatusInternalServerError)
		return
	}

	// 4. Serializamos a resposta ANTES de escrever qualquer header.
	//
	// Por que esta ordem importa?
	// Em HTTP, o status code e os headers são enviados ao cliente assim que
	// qualquer escrita acontece no ResponseWriter. Se chamarmos w.WriteHeader(200)
	// e depois json.Marshal falhar, já é tarde demais para mudar o status para 500 —
	// o cliente recebeu um 200 com corpo vazio ou incompleto.
	//
	// A solução: converter para []byte primeiro. Se isso falhar, ainda podemos
	// responder com 500. Só então escrevemos o status e o corpo.
	//
	// Analogia JS: é como fazer `const body = JSON.stringify(data)` antes de
	// `res.status(200).send(body)` — evitar chamar res.status() cedo demais.
	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta de login", "detalhe", err)
		http.Error(w, "erro interno no servidor", http.StatusInternalServerError)
		return
	}

	// 5. Tudo certo — agora sim escrevemos os headers e o corpo de uma vez
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(corpo) //nolint:errcheck // Erro de escrita de rede após WriteHeader não é acionável
}

