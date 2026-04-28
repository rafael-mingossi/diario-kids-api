package models

import (
	"time" //pacote nativo do Go para lidar com datas e horas

	"gorm.io/gorm"
)

// Usuario representa Pais, Professores, Diretores e Proprietários.
type Usuario struct {
	// gorm.Model cria sozinho: ID, CreatedAt, UpdatedAt e DeletedAt
	gorm.Model
	Nome      string
	Email     string `gorm:"uniqueIndex;not null"` // Garante que o email não se repita
	SenhaHash string `gorm:"not null"`
	Role      string `gorm:"not null"`

	// Relação Múltipla: Um pai tem vários alunos, um aluno tem vários pais.
	// O GORM vai criar a tabela 'aluno_pais' sozinho por causa dessa tag 'many2many'
	Alunos []Aluno `gorm:"many2many:aluno_pais;"`
}

// Sala representa o Berçário, Infantil 1, etc.
type Sala struct {
	gorm.Model
	Nome string `gorm:"not null"`

	// Relação 1:1 com Professor (Uma sala tem 1 professor)
	ProfessorID uint
	Professor   Usuario `gorm:"foreignKey:ProfessorID"`

	// Relação 1:N com Alunos (Uma sala tem muitos alunos)
	Alunos []Aluno
}

// Aluno representa as crianças
type Aluno struct {
	gorm.Model
	Nome string `gorm:"not null"`
	// NOVIDADE: Alterado de string para time.Time
	DataNascimento time.Time `gorm:"type:date;not null"`

	// Relação de Pertencimento: O aluno pertence a uma sala
	SalaID uint
	Sala   Sala

	// Relação de volta para os pais
	Pais []Usuario `gorm:"many2many:aluno_pais;"`
}
