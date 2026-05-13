package services

import (
	"encoding/json"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
)

type AuditService interface {
	Registrar(input AuditLogInput) error
}

type AuditLogInput struct {
	ActorUserID       *uint
	ActorPlatformRole string
	ActorSchoolRole   string
	ActorEscolaID     *uint
	TargetEscolaID    *uint
	Action            string
	EntityType        string
	EntityID          string
	Origin            string
	Before            any
	After             any
	IP                string
	UserAgent         string
}

type auditService struct {
	repo repository.AuditLogRepository
}

func NewAuditService(repo repository.AuditLogRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Registrar(input AuditLogInput) error {
	beforeJSON, err := marshalAuditPayload(input.Before)
	if err != nil {
		return fmt.Errorf("erro ao serializar before_json da auditoria: %w", err)
	}

	afterJSON, err := marshalAuditPayload(input.After)
	if err != nil {
		return fmt.Errorf("erro ao serializar after_json da auditoria: %w", err)
	}

	registro := models.AuditLog{
		ActorUserID:       input.ActorUserID,
		ActorPlatformRole: input.ActorPlatformRole,
		ActorSchoolRole:   input.ActorSchoolRole,
		ActorEscolaID:     input.ActorEscolaID,
		TargetEscolaID:    input.TargetEscolaID,
		Action:            input.Action,
		EntityType:        input.EntityType,
		EntityID:          input.EntityID,
		Origin:            input.Origin,
		BeforeJSON:        beforeJSON,
		AfterJSON:         afterJSON,
		IP:                input.IP,
		UserAgent:         input.UserAgent,
	}

	if err := s.repo.Criar(&registro); err != nil {
		return fmt.Errorf("erro ao persistir auditoria: %w", err)
	}

	return nil
}

func marshalAuditPayload(payload any) (json.RawMessage, error) {
	if payload == nil {
		return nil, nil
	}

	corpo, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(corpo), nil
}
