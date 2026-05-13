package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

type ClienteRepository interface {
	Criar(cliente *models.Cliente) error
	BuscarPorID(id uint) (*models.Cliente, error)
	BuscarPorDocumento(documento string) (*models.Cliente, error)
}

type clienteRepository struct {
	db *gorm.DB
}

func NewClienteRepository(db *gorm.DB) ClienteRepository {
	return &clienteRepository{db: db}
}

var ErrClienteDuplicadoDB = errors.New("violação de restrição única (cliente já existe)")

func (r *clienteRepository) Criar(cliente *models.Cliente) error {
	err := r.db.Create(cliente).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrClienteDuplicadoDB
		}
		return err
	}
	return nil
}

func (r *clienteRepository) BuscarPorID(id uint) (*models.Cliente, error) {
	var cliente models.Cliente
	err := r.db.First(&cliente, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cliente, nil
}

func (r *clienteRepository) BuscarPorDocumento(documento string) (*models.Cliente, error) {
	var cliente models.Cliente
	err := r.db.Where("documento = ?", documento).First(&cliente).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cliente, nil
}
