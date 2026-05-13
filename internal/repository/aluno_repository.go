package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

// interface pública porque o Service depende dela
type AlunoRepository interface {
	Criar(aluno *models.Aluno) error
	BuscarPorMatricula(matricula string) (*models.Aluno, error)
}

// interface privada
type alunoRepository struct {
	db *gorm.DB
}

// constructor
func NewAlunoRepository(db *gorm.DB) AlunoRepository {
	return &alunoRepository{db: db}
}

// Constante pública para o Service saber o que aconteceu, sem acoplar ao detalhe do Postgres.
var ErrAlunoDuplicadoDB = errors.New("violação de restrição única (aluno já existe)")

func (r *alunoRepository) Criar(aluno *models.Aluno) error {
	err := r.db.Create(aluno).Error

	if err != nil {
		// Se o banco acusar conflito de unicidade (23505), devolvemos uma sentinel.
		// O service decide o que fazer com isso sem precisar conhecer pgconn/SQLSTATE.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlunoDuplicadoDB
		}
		return err
	}
	return nil
}

func (r *alunoRepository) BuscarPorMatricula(matricula string) (*models.Aluno, error) {
	var aluno models.Aluno
	err := r.db.Where("matricula = ?", matricula).First(&aluno).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &aluno, nil
}
