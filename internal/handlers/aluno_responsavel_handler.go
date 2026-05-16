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

type AlunoResponsavelHandler struct {
	service  services.AlunoResponsavelService
	audit    services.AuditService
	validate *validator.Validate
}

func NewAlunoResponsavelHandler(service services.AlunoResponsavelService, audit services.AuditService) *AlunoResponsavelHandler {
	return &AlunoResponsavelHandler{service: service, audit: audit, validate: validator.New()}
}

func (h *AlunoResponsavelHandler) Vincular(w http.ResponseWriter, r *http.Request) {
	var input dto.VincularAlunoResponsavelInput

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

	resposta, err := h.service.Vincular(input)
	if err != nil {
		if errors.Is(err, services.ErrAlunoNaoEncontrado) {
			writeJSONError(w, http.StatusUnprocessableEntity, "aluno informado não existe")
			return
		}
		if errors.Is(err, services.ErrResponsavelNaoEncontrado) {
			writeJSONError(w, http.StatusUnprocessableEntity, "usuário responsável informado não existe")
			return
		}
		if errors.Is(err, services.ErrResponsavelNaoPertenceAEscola) {
			writeJSONError(w, http.StatusUnprocessableEntity, "responsável não pertence à escola do aluno")
			return
		}
		if errors.Is(err, services.ErrUsuarioNaoEhResponsavel) {
			writeJSONError(w, http.StatusUnprocessableEntity, "usuário informado não tem papel de responsável")
			return
		}
		if errors.Is(err, services.ErrAlunoResponsavelDuplicado) {
			writeJSONError(w, http.StatusConflict, "responsável já vinculado a este aluno")
			return
		}

		slog.Error("erro interno ao vincular aluno a responsável", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno no servidor")
		return
	}

	actorUserID, actorPlatformRole, actorSchoolRole, actorEscolaID := actorFromRequest(r)
	targetEscolaID := resposta.EscolaID
	registrarAuditoriaBestEffort(h.audit, r, services.AuditLogInput{
		ActorUserID:       actorUserID,
		ActorPlatformRole: actorPlatformRole,
		ActorSchoolRole:   actorSchoolRole,
		ActorEscolaID:     actorEscolaID,
		TargetEscolaID:    &targetEscolaID,
		Action:            "link_user_student",
		EntityType:        "aluno_responsavel",
		EntityID:          strconv.FormatUint(uint64(resposta.ID), 10),
		Origin:            "api",
		After:             resposta,
	})

	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta de vínculo aluno-responsável", "detalhe", err)
		writeJSONError(w, http.StatusInternalServerError, "erro interno no servidor")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(corpo) //nolint:errcheck
}
