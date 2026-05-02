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
	// Nome pode se repetir entre salas diferentes.
	// A unicidade real da sala será a combinação Nome + Numero.
	Nome string `gorm:"not null;uniqueIndex:idx_sala_nome_numero"`

	// Numero identifica a turma/sala dentro do mesmo Nome.
	// Exemplos válidos de negócio: "1", "2", "A", "B".
	Numero string `gorm:"not null;uniqueIndex:idx_sala_nome_numero"`

	// Professor é opcional no momento da criação da sala.
	// Usamos *uint para representar ausência real (nil), em vez de 0.
	ProfessorID *uint
	Professor   *Usuario `gorm:"foreignKey:ProfessorID"`

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
