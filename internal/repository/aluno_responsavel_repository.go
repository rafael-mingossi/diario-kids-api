package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

type AlunoResponsavelRepository interface {
	Criar(vinculo *models.AlunoResponsavel) error
	BuscarPorAlunoEUsuario(alunoID uint, usuarioID uint) (*models.AlunoResponsavel, error)
}

type alunoResponsavelRepository struct {
	db *gorm.DB
}

func NewAlunoResponsavelRepository(db *gorm.DB) AlunoResponsavelRepository {
	return &alunoResponsavelRepository{db: db}
}

var ErrAlunoResponsavelDuplicadoDB = errors.New("violação de restrição única (vínculo aluno-responsável duplicado)")

func (r *alunoResponsavelRepository) Criar(vinculo *models.AlunoResponsavel) error {
	err := r.db.Create(vinculo).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlunoResponsavelDuplicadoDB
		}
		return err
	}
	return nil
}

func (r *alunoResponsavelRepository) BuscarPorAlunoEUsuario(alunoID uint, usuarioID uint) (*models.AlunoResponsavel, error) {
	var vinculo models.AlunoResponsavel
	err := r.db.Where("aluno_id = ? AND usuario_id = ?", alunoID, usuarioID).First(&vinculo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vinculo, nil
}
