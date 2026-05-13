package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

type UsuarioEscolaRepository interface {
	Criar(vinculo *models.UsuarioEscola) error
	BuscarPorUsuarioEEscola(usuarioID uint, escolaID uint) (*models.UsuarioEscola, error)
}

type usuarioEscolaRepository struct {
	db *gorm.DB
}

func NewUsuarioEscolaRepository(db *gorm.DB) UsuarioEscolaRepository {
	return &usuarioEscolaRepository{db: db}
}

var ErrUsuarioEscolaDuplicadoDB = errors.New("violação de restrição única (vínculo usuário-escola duplicado)")

func (r *usuarioEscolaRepository) Criar(vinculo *models.UsuarioEscola) error {
	err := r.db.Create(vinculo).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUsuarioEscolaDuplicadoDB
		}
		return err
	}
	return nil
}

func (r *usuarioEscolaRepository) BuscarPorUsuarioEEscola(usuarioID uint, escolaID uint) (*models.UsuarioEscola, error) {
	var vinculo models.UsuarioEscola
	err := r.db.Where("usuario_id = ? AND escola_id = ?", usuarioID, escolaID).First(&vinculo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vinculo, nil
}