package repository

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Criar(log *models.AuditLog) error
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Criar(log *models.AuditLog) error {
	return r.db.Create(log).Error
}
