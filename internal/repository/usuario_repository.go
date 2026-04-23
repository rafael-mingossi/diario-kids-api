package repository

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

// UsuarioRepository é a "classe" que vai cuidar dos usuários no banco
type UsuarioRepository struct {
	db *gorm.DB
}

// NewUsuarioRepository é como se fosse o construtor da nossa classe
func NewUsuarioRepository(db *gorm.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

// CriarUsuario recebe um modelo de usuário e salva no banco
func (r *UsuarioRepository) CriarUsuario(usuario *models.Usuario) error {
	// O GORM faz o INSERT no Postgres aqui. Se der erro (ex: email já existe), ele retorna.
	resultado := r.db.Create(usuario)
	return resultado.Error
}
