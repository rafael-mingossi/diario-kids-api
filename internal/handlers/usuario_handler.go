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

type UsuarioHandler struct {
	// NOVIDADE: Agora usamos a Interface (sem o asterisco *)
	service  services.UsuarioService
	validate *validator.Validate
}

// Recebe o Service e inicializa o Validador
func NewUsuarioHandler(service services.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *UsuarioHandler) CriarUsuario(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarUsuarioInput

	// 1. Lê o JSON (O Garçom anota o pedido)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON mal formatado", http.StatusBadRequest)
		return
	}

	// 2. Valida o input automaticamente usando as tags do DTO
	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "Dados de entrada inválidos (verifique email, tamanho da senha, etc)", http.StatusUnprocessableEntity)
		return
	}

	// 3. Manda para a Cozinha (Service) fazer o trabalho pesado
	resposta, err := h.service.CriarUsuario(input)
	if err != nil {
		// Feedback 2: Tratamos o erro específico de conflito (HTTP 409)
		if errors.Is(err, services.ErrEmailEmUso) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		// Feedback 1: Segurança (OWASP) - Escondemos o erro real do usuário, logamos no terminal
		slog.Error("Erro interno ao criar usuário", "detalhe", err)
		http.Error(w, "Erro interno no servidor. Tente novamente mais tarde.", http.StatusInternalServerError)
		return
	}

	// 4. Entrega o prato (HTTP 201 Created com a resposta limpa)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Feedback 4: Não ignoramos o erro do JSON Encoder
	if err := json.NewEncoder(w).Encode(resposta); err != nil {
		slog.Error("Erro ao enviar resposta JSON", "detalhe", err)
	}
}
