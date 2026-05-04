package repository

import (
	// "errors"

	// Pacote do motor do Postgres
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

// interface pública porque o Service depende dela
type AlunoRepository interface {
	Criar(aluno *models.Aluno) error
}

// interface privada
type alunoRepository struct {
	db *gorm.DB
}

// constructor
func NewAlunoRepository(db *gorm.DB) AlunoRepository {
	return &alunoRepository{db: db}
}

// Constante publica para o Service saber o que aconteceu, sem saber de Postgres
// var ErrAlunoDuplicadoDB = errors.New("violação de restrição única (aluno já existe)")

func (r *alunoRepository) Criar(aluno *models.Aluno) error {
	err := r.db.Create(aluno).Error

	if err != nil {
		// //verificar se o erro é especifico do postgres
		// var pgErr *pgconn.PgError
		// // Se for um erro do Postgres E o código for 23505 (Unique Violation)
		// if errors.As(err, &pgErr) && pgErr.Code == "23505"{
		// 	return ErrAlunoDuplicadoDB
		// }
		// // se for outro erro
		return err
	}
	return nil
}

// TODO criar funcao de buscar aluno por matricula
