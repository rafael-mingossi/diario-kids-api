package handlers

import (
	"encoding/json"
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

	// 1. Lê o JSON
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON mal formatado", http.StatusBadRequest)
		return
	}

	// 2. Valida o input automaticamente usando as tags do DTO
	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "Dados de entrada inválidos (verifique email, tamanho da senha, etc)", http.StatusUnprocessableEntity)
		return
	}

	// 3. Manda para o Service fazer o trabalho pesado
	resposta, err := h.service.Login(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 4. Devolve a resposta do Service como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Explicita que deu 200 OK

	if err := json.NewEncoder(w).Encode(resposta); err != nil {
		// Ignorar esse erro pode esconder problemas de rede, igual vimos antes
		http.Error(w, "Erro ao serializar resposta", http.StatusInternalServerError)
	}
}
