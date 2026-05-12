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

type AlunoHandler struct {
	service  services.AlunoService
	validate *validator.Validate
}

// construtor, recebe o service e inicializa o validador
func NewAlunoHandler(service services.AlunoService) *AlunoHandler {
	return &AlunoHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *AlunoHandler) CriarAluno(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarAlunoInput

	//1. limita o tamanho do body
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) //1MB

	//2. le o JSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "corpo da requisição muito grande")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "JSON mal formatado")
		return
	}

	//3. valida o input usando DTO tags
	if err := h.validate.Struct(input); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "dados de entrada inválidos")
		return
	}

	//4. manda pro Service
	resposta, err := h.service.CriarAluno(input)
	if err != nil {
		// Erros de domínio esperados viram 422. Não são falhas internas do servidor.
		if errors.Is(err, services.ErrDataNascimentoInvalida) {
			writeJSONError(w, http.StatusUnprocessableEntity, "data_nascimento inválida")
			return
		}
		if errors.Is(err, services.ErrDataNascimentoFutura) {
			writeJSONError(w, http.StatusUnprocessableEntity, "data_nascimento não pode ser futura")
			return
		}
		if errors.Is(err, services.ErrSalaNaoEncontrada) {
			// O cliente enviou um sala_id, mas essa sala não existe no banco.
			// Isso é erro de entrada/negócio, não erro interno do servidor.
			writeJSONError(w, http.StatusUnprocessableEntity, "sala_id inválido")
			return
		}
		if errors.Is(err, services.ErrEscolaNaoEncontrada) {
			writeJSONError(w, http.StatusUnprocessableEntity, "escola_id inválido")
			return
		}
		if errors.Is(err, services.ErrSalaNaoPertenceAEscola) {
			writeJSONError(w, http.StatusUnprocessableEntity, "sala_id não pertence à escola informada")
			return
		}
		//A: Segurança (OWASP) - Escondemos o erro real do usuário, logamos no terminal
		slog.Error("Erro interno ao criar aluno", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno no servidor")
		return
	}

	// 4. Serializamos a resposta ANTES de escrever qualquer header.
	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("Erro ao serializar resposta", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno no servidor")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(corpo)
}
