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

type SalaHandler struct {
	service  services.SalaService
	validate *validator.Validate
}

// recebe o service e inicializa o validador e retorna um ponteiro para o handler
func NewSalaHandler(service services.SalaService) *SalaHandler {
	return &SalaHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *SalaHandler) CriarSala(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarSalaInput

	// 1. Limita o tamanho do body para prevenir "body bomb" (DoS por payload gigante).
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB

	// 2. Lê o JSON do corpo da requisição.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "corpo da requisição muito grande", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "JSON mal formatado", http.StatusBadRequest)
		return
	}

	// 3. Valida os campos usando as tags do DTO (required, etc.)
	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "dados de entrada inválidos", http.StatusUnprocessableEntity)
		return
	}

	// 4. Delega ao Service — ele verifica se a sala já existe, insere no banco e retorna o DTO de resposta.
	resposta, err := h.service.CriarSala(input)
	if err != nil {
		if errors.Is(err, services.ErrSalaDuplicada) {
			http.Error(w, "esta sala já existe", http.StatusConflict)
			return
		}
		// Erro interno inesperado: logamos o detalhe no servidor, nunca no cliente.
		slog.Error("erro interno ao criar sala",
			"path", r.URL.Path,
			"ip", r.RemoteAddr,
			"detalhe", err,
		)
		http.Error(w, "erro interno do servidor", http.StatusInternalServerError)
		return
	}

	// 5. Serializamos ANTES de escrever qualquer header.
	// Se o Marshal falhar depois do WriteHeader(201), o cliente receberia 201 com corpo vazio.
	// Serializando primeiro, ainda podemos retornar 500 se algo der errado.
	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta de sala", "detalhe", err)
		http.Error(w, "erro interno do servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201: recurso criado com sucesso
	w.Write(corpo)                    //nolint:errcheck
}
