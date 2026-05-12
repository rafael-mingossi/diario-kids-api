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

// Escola representa uma unidade organizacional do sistema.
// A mesma tabela serve tanto para matriz quanto para filial.
// Quando IsMatriz=true, MatrizID fica nil.
// Quando IsMatriz=false, MatrizID aponta para a escola matriz dona da filial.
type Escola struct {
	gorm.Model
	CNPJ         string `gorm:"uniqueIndex;not null"`
	RazaoSocial  string `gorm:"not null"`
	NomeFantasia string `gorm:"not null"`
	IsMatriz     bool   `gorm:"not null;default:false"`

	MatrizID *uint
	Matriz   *Escola  `gorm:"foreignKey:MatrizID"`
	Filiais  []Escola `gorm:"foreignKey:MatrizID"`

	// Relações operacionais futuras já nascem escopadas por unidade.
	Salas  []Sala
	Alunos []Aluno
}

// Sala representa o Berçário, Infantil 1, etc.
type Sala struct {
	gorm.Model
	// Transitional note:
	// já existem salas legadas no banco sem escola associada.
	// Por isso o campo fica nullable no schema por enquanto, para a migration não falhar.
	// A API continua exigindo escola_id nos novos cadastros via DTO + service.
	EscolaID *uint   `gorm:"uniqueIndex:idx_sala_escola_nome_numero"`
	Escola   *Escola `gorm:"foreignKey:EscolaID"`

	// Nome pode se repetir entre salas diferentes.
	// A unicidade real da sala passa a ser EscolaID + Nome + Numero.
	Nome string `gorm:"not null;uniqueIndex:idx_sala_escola_nome_numero"`

	// Numero identifica a turma/sala dentro do mesmo Nome.
	// Exemplos válidos de negócio: "1", "2", "A", "B".
	Numero string `gorm:"not null;uniqueIndex:idx_sala_escola_nome_numero"`

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
	// Transitional note:
	// existem alunos legados sem escola associada no banco atual.
	// Mantemos nullable no schema para permitir a migration, mas a API exige escola_id
	// para qualquer novo aluno criado a partir desta fase.
	EscolaID *uint   `gorm:"index"`
	Escola   *Escola `gorm:"foreignKey:EscolaID"`

	Nome string `gorm:"not null"`
	// NOVIDADE: Alterado de string para time.Time
	DataNascimento time.Time `gorm:"type:date;not null"`
	Matricula      string    `gorm:"not null;unique"`

	// Relação de Pertencimento: O aluno pertence a uma sala
	SalaID *uint
	Sala   *Sala

	// Relação de volta para os pais
	Pais []Usuario `gorm:"many2many:aluno_pais;"`
}
