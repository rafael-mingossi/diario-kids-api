package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/services"
)

type ClienteHandler struct {
	service  services.ClienteService
	audit    services.AuditService
	validate *validator.Validate
}

func NewClienteHandler(service services.ClienteService, audit services.AuditService) *ClienteHandler {
	return &ClienteHandler{service: service, audit: audit, validate: validator.New()}
}

func (h *ClienteHandler) CriarCliente(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarClienteInput

	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024)
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

	if err := h.validate.Struct(input); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "dados de entrada inválidos")
		return
	}

	resposta, err := h.service.CriarCliente(input)
	if err != nil {
		if errors.Is(err, services.ErrClienteDuplicado) {
			writeJSONError(w, http.StatusConflict, "cliente já existe")
			return
		}

		slog.Error("erro interno ao criar cliente",
			"path", r.URL.Path,
			"ip", r.RemoteAddr,
			"detalhe", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "erro interno do servidor")
		return
	}

	actorUserID, actorPlatformRole, actorSchoolRole, actorEscolaID := actorFromRequest(r)
	registrarAuditoriaBestEffort(h.audit, r, services.AuditLogInput{
		ActorUserID:       actorUserID,
		ActorPlatformRole: actorPlatformRole,
		ActorSchoolRole:   actorSchoolRole,
		ActorEscolaID:     actorEscolaID,
		Action:            "create",
		EntityType:        "cliente",
		EntityID:          strconv.FormatUint(uint64(resposta.ID), 10),
		Origin:            "api",
		After:             resposta,
	})

	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta de cliente", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno do servidor")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(corpo) //nolint:errcheck
}
