package services

import (
	"errors"
	"fmt"

	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SetupService interface {
	SetupInicial(input dto.SetupInicialInput) (*dto.SetupInicialResponse, error)
}

type setupService struct {
	db *gorm.DB
}

func NewSetupService(db *gorm.DB) SetupService {
	return &setupService{db: db}
}

var ErrSistemaJaInicializado = errors.New("sistema já inicializado")

func (s *setupService) SetupInicial(input dto.SetupInicialInput) (*dto.SetupInicialResponse, error) {
	var totalUsuarios int64
	if err := s.db.Model(&models.Usuario{}).Count(&totalUsuarios).Error; err != nil {
		return nil, fmt.Errorf("erro ao contar usuários: %w", err)
	}

	var totalEscolas int64
	if err := s.db.Model(&models.Escola{}).Count(&totalEscolas).Error; err != nil {
		return nil, fmt.Errorf("erro ao contar escolas: %w", err)
	}

	if totalUsuarios > 0 || totalEscolas > 0 {
		return nil, ErrSistemaJaInicializado
	}

	senhaHash, err := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao criptografar senha: %w", err)
	}

	var usuario models.Usuario

	err = s.db.Transaction(func(tx *gorm.DB) error {
		usuario = models.Usuario{
			Nome:         input.Nome,
			Email:        input.Email,
			SenhaHash:    string(senhaHash),
			PlatformRole: "platform_admin",
		}
		if err := tx.Create(&usuario).Error; err != nil {
			return fmt.Errorf("erro ao criar usuário inicial: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	token, err := gerarJWT(usuario.ID, usuario.Email, "", nil, usuario.PlatformRole)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar token inicial: %w", err)
	}

	return &dto.SetupInicialResponse{
		UsuarioID:    usuario.ID,
		Email:        usuario.Email,
		PlatformRole: usuario.PlatformRole,
		Token:        token,
	}, nil
}
