package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn" // Pacote do motor do Postgres
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

// UsuarioRepository é a "classe" que vai cuidar dos usuários no banco
// 1. O Contrato (Interface)
type UsuarioRepository interface {
	Criar(usuario *models.Usuario) error
	BuscarPorEmail(email string) (*models.Usuario, error)
}

// 2. A Implementação Privada
type usuarioRepository struct {
	db *gorm.DB
}

// 3. O Construtor que devolve a Interface
func NewUsuarioRepository(db *gorm.DB) UsuarioRepository {
	return &usuarioRepository{db: db}
}

// NOVIDADE: Uma constante pública para o Service saber o que aconteceu, sem saber de Postgres
var ErrEmailDuplicadoDB = errors.New("violação de restrição única (email duplicado)")

func (r *usuarioRepository) Criar(usuario *models.Usuario) error {
	err := r.db.Create(usuario).Error
	if err != nil {
		// Verificamos se o erro é específico do Postgres (Driver pgx)
		var pgErr *pgconn.PgError
		// Se for um erro do Postgres E o código for 23505 (Unique Violation)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrEmailDuplicadoDB
		}
		// Se for outro erro (cabo de rede solto, etc)
		return err
	}
	return nil
}

// Método para checar duplicidade (e futuro login)
func (r *usuarioRepository) BuscarPorEmail(email string) (*models.Usuario, error) {
	var usuario models.Usuario

	// Busca o primeiro registro onde o email bata
	err := r.db.Where("email = ?", email).First(&usuario).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Não é um erro fatal, apenas significa que o usuário não existe
		}
		return nil, err // Erro real de banco (banco caiu, etc)
	}

	return &usuario, nil
}
