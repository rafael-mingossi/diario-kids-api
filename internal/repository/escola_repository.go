package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

type EscolaRepository interface {
	Criar(escola *models.Escola) error
	BuscarPorID(id uint) (*models.Escola, error)
	BuscarPorCNPJ(cnpj string) (*models.Escola, error)
}

type escolaRepository struct {
	db *gorm.DB
}

func NewEscolaRepository(db *gorm.DB) EscolaRepository {
	return &escolaRepository{db: db}
}

var ErrEscolaDuplicadaDB = errors.New("violação de restrição única (escola já existe)")

func (r *escolaRepository) Criar(escola *models.Escola) error {
	err := r.db.Create(escola).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrEscolaDuplicadaDB
		}
		return err
	}
	return nil
}

func (r *escolaRepository) BuscarPorID(id uint) (*models.Escola, error) {
	var escola models.Escola
	err := r.db.First(&escola, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &escola, nil
}

func (r *escolaRepository) BuscarPorCNPJ(cnpj string) (*models.Escola, error) {
	var escola models.Escola
	err := r.db.Where("cnpj = ?", cnpj).First(&escola).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &escola, nil
}