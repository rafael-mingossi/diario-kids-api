package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn" // Pacote do motor do Postgres
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"gorm.io/gorm"
)

// interface pública porque o Service depende dela
type SalaRepository interface {
	Criar(sala *models.Sala) error
	BuscarPorNomeENumero(nome string, numero string) (*models.Sala, error)
}

// interface privada
type salaRepository struct {
	db *gorm.DB
}

// constructor
func NewSalaRepository(db *gorm.DB) SalaRepository {
	return &salaRepository{db: db}
}

// Constante publica para o Service saber o que aconteceu, sem saber de Postgres
var ErrSalaDuplicadaDB = errors.New("violação de restrição única (sala já existe)")

func (r *salaRepository) Criar(sala *models.Sala) error {
	err := r.db.Create(sala).Error
	if err != nil {
		// Verificamos se o erro é específico do Postgres
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSalaDuplicadaDB
		}
		return err
	}
	return nil
}

func (r *salaRepository) BuscarPorNomeENumero(nome string, numero string) (*models.Sala, error) {
	var sala models.Sala

	// A regra de negócio diz que a combinação Nome + Numero identifica a sala.
	// Exemplo: "Infantil 2" + "A" pode existir, mas outra "Infantil 2" + "A" não.
	err := r.db.Where("nome = ? AND numero = ?", nome, numero).First(&sala).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Não encontrou: não é erro fatal, só significa "sala ainda não existe"
		}
		return nil, err // Erro real de banco
	}

	return &sala, nil
}
