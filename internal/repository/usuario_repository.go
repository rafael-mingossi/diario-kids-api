package repository

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

// UsuarioRepository é a "classe" que vai cuidar dos usuários no banco
// 1. O Contrato (Interface)
type UsuarioRepository interface {
	Criar(usuario *models.Usuario) error
	// No futuro adicionaremos: BuscarPorEmail(email string) (*models.Usuario, error)
}

// 2. A Implementação Privada
type usuarioRepository struct {
	db *gorm.DB
}

// 3. O Construtor que devolve a Interface
func NewUsuarioRepository(db *gorm.DB) UsuarioRepository {
	return &usuarioRepository{db: db}
}

func (r *usuarioRepository) Criar(usuario *models.Usuario) error {
	return r.db.Create(usuario).Error
}
