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

type EscolaHandler struct {
	service  services.EscolaService
	audit    services.AuditService
	validate *validator.Validate
}

func NewEscolaHandler(service services.EscolaService, audit services.AuditService) *EscolaHandler {
	return &EscolaHandler{
		service:  service,
		audit:    audit,
		validate: validator.New(),
	}
}

func (h *EscolaHandler) CriarEscola(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarEscolaInput

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

	resposta, err := h.service.CriarEscola(input)
	if err != nil {
		if errors.Is(err, services.ErrEscolaDuplicada) {
			writeJSONError(w, http.StatusConflict, "escola já existe")
			return
		}
		if errors.Is(err, services.ErrClienteDaEscolaNaoEncontrado) {
			writeJSONError(w, http.StatusUnprocessableEntity, "cliente informado não existe")
			return
		}
		if errors.Is(err, services.ErrMatrizObrigatoria) || errors.Is(err, services.ErrMatrizInvalida) || errors.Is(err, services.ErrMatrizNaoPodeTerMatriz) {
			writeJSONError(w, http.StatusUnprocessableEntity, "dados da hierarquia matriz/filial inválidos")
			return
		}
		if errors.Is(err, services.ErrMatrizDeOutroCliente) {
			writeJSONError(w, http.StatusUnprocessableEntity, "matriz informada pertence a outro cliente")
			return
		}

		slog.Error("erro interno ao criar escola",
			"path", r.URL.Path,
			"ip", r.RemoteAddr,
			"detalhe", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "erro interno do servidor")
		return
	}

	actorUserID, actorPlatformRole, actorSchoolRole, actorEscolaID := actorFromRequest(r)
	targetEscolaID := resposta.ID
	registrarAuditoriaBestEffort(h.audit, r, services.AuditLogInput{
		ActorUserID:       actorUserID,
		ActorPlatformRole: actorPlatformRole,
		ActorSchoolRole:   actorSchoolRole,
		ActorEscolaID:     actorEscolaID,
		TargetEscolaID:    &targetEscolaID,
		Action:            "create",
		EntityType:        "escola",
		EntityID:          strconv.FormatUint(uint64(resposta.ID), 10),
		Origin:            "api",
		After:             resposta,
	})

	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta de escola", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno do servidor")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(corpo) //nolint:errcheck
}
