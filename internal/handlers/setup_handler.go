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

type SetupHandler struct {
	service  services.SetupService
	audit    services.AuditService
	validate *validator.Validate
}

func NewSetupHandler(service services.SetupService, audit services.AuditService) *SetupHandler {
	return &SetupHandler{service: service, audit: audit, validate: validator.New()}
}

func (h *SetupHandler) SetupInicial(w http.ResponseWriter, r *http.Request) {
	var input dto.SetupInicialInput

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

	resposta, err := h.service.SetupInicial(input)
	if err != nil {
		if errors.Is(err, services.ErrSistemaJaInicializado) {
			writeJSONError(w, http.StatusConflict, "sistema já foi inicializado")
			return
		}

		slog.Error("erro interno no setup inicial",
			"path", r.URL.Path,
			"ip", r.RemoteAddr,
			"detalhe", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "erro interno no servidor")
		return
	}

	registrarAuditoriaBestEffort(h.audit, r, services.AuditLogInput{
		Action:            "setup_initial",
		EntityType:        "usuario",
		EntityID:          strconv.FormatUint(uint64(resposta.UsuarioID), 10),
		Origin:            "setup_inicial",
		ActorUserID:       &resposta.UsuarioID,
		ActorPlatformRole: resposta.PlatformRole,
		After:             resposta,
	})

	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta do setup inicial", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno no servidor")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(corpo) //nolint:errcheck
}
