package models

import (
	"encoding/json"
	"time" //pacote nativo do Go para lidar com datas e horas

	"gorm.io/gorm"
)

// Usuario representa identidades de login da plataforma.
// O papel operacional dele por escola fica em UsuarioEscola.
type Usuario struct {
	// gorm.Model cria sozinho: ID, CreatedAt, UpdatedAt e DeletedAt
	gorm.Model
	Nome      string
	Email     string `gorm:"uniqueIndex;not null"` // Garante que o email não se repita
	SenhaHash string `gorm:"not null"`
	// PlatformRole representa permissões globais do SaaS e não substitui
	// os papéis operacionais da escola.
	PlatformRole string `gorm:"index"`

	UsuarioEscolas []UsuarioEscola

	// Um mesmo usuário responsável pode estar ligado a vários alunos.
	AlunoResponsaveis []AlunoResponsavel
}

// Cliente representa a conta comercial do SaaS.
// É aqui que vivem dados contratuais, de cobrança e status da conta.
// Um cliente pode ter uma ou várias unidades operacionais (Escolas).
type Cliente struct {
	gorm.Model
	Nome            string `gorm:"not null"`
	Documento       string `gorm:"uniqueIndex;not null"`
	EmailFinanceiro string
	Telefone        string
	Status          string `gorm:"not null;default:ativo"`
	Plano           string `gorm:"not null;default:basico"`

	Escolas []Escola
}

// Escola representa uma unidade organizacional do sistema.
// A mesma tabela serve tanto para matriz quanto para filial.
// Quando IsMatriz=true, MatrizID fica nil.
// Quando IsMatriz=false, MatrizID aponta para a escola matriz dona da filial.
type Escola struct {
	gorm.Model
	ClienteID    uint     `gorm:"not null;index"`
	Cliente      *Cliente `gorm:"foreignKey:ClienteID"`
	CNPJ         string   `gorm:"uniqueIndex;not null"`
	RazaoSocial  string   `gorm:"not null"`
	NomeFantasia string   `gorm:"not null"`
	IsMatriz     bool     `gorm:"not null;default:false"`

	MatrizID *uint
	Matriz   *Escola  `gorm:"foreignKey:MatrizID"`
	Filiais  []Escola `gorm:"foreignKey:MatrizID"`

	// Relações operacionais futuras já nascem escopadas por unidade.
	Salas          []Sala
	Alunos         []Aluno
	UsuarioEscolas []UsuarioEscola
}

// UsuarioEscola representa o vínculo do usuário com uma unidade específica.
// É aqui que o papel começa a ser contextualizado por escola/filial.
// Exemplo real de negócio:
// - usuário A pode ser coordenador da filial 1
// - usuário B pode ser proprietário da matriz
// - no futuro, o mesmo usuário poderia até ter papéis diferentes em unidades diferentes
type UsuarioEscola struct {
	gorm.Model
	UsuarioID uint     `gorm:"not null;uniqueIndex:idx_usuario_escola"`
	Usuario   *Usuario `gorm:"foreignKey:UsuarioID"`
	EscolaID  uint     `gorm:"not null;uniqueIndex:idx_usuario_escola"`
	Escola    *Escola  `gorm:"foreignKey:EscolaID"`
	Role      string   `gorm:"not null"`
	Ativo     bool     `gorm:"not null;default:true"`
}

// Sala representa o Berçário, Infantil 1, etc.
type Sala struct {
	gorm.Model
	EscolaID uint    `gorm:"not null;uniqueIndex:idx_sala_escola_nome_numero"`
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
	EscolaID uint    `gorm:"not null;index"`
	Escola   *Escola `gorm:"foreignKey:EscolaID"`

	Nome string `gorm:"not null"`
	// NOVIDADE: Alterado de string para time.Time
	DataNascimento time.Time `gorm:"type:date;not null"`
	Matricula      string    `gorm:"not null;unique"`

	// Relação de Pertencimento: O aluno pertence a uma sala
	SalaID *uint
	Sala   *Sala

	// Relação de volta para os responsáveis.
	AlunoResponsaveis []AlunoResponsavel
}

// AlunoResponsavel guarda o vínculo explícito entre um aluno e um usuário responsável.
// Aqui ficam os metadados de negócio que não pertencem nem ao usuário nem ao aluno isoladamente.
// Exemplos:
// - parentesco: mae, pai, avo, responsavel_legal, outro
// - pode receber notificações
// - pode buscar a criança
// - é contato de emergência
type AlunoResponsavel struct {
	gorm.Model
	AlunoID            uint     `gorm:"not null;uniqueIndex:idx_aluno_responsavel"`
	Aluno              *Aluno   `gorm:"foreignKey:AlunoID"`
	UsuarioID          uint     `gorm:"not null;uniqueIndex:idx_aluno_responsavel"`
	Usuario            *Usuario `gorm:"foreignKey:UsuarioID"`
	Parentesco         string   `gorm:"not null"`
	ResponsavelLegal   bool     `gorm:"not null;default:false"`
	RecebeNotificacoes bool     `gorm:"not null;default:true"`
	ContatoEmergencia  bool     `gorm:"not null;default:false"`
	AutorizadoBusca    bool     `gorm:"not null;default:false"`
	Observacao         string
	Ativo              bool `gorm:"not null;default:true"`
}

func (AlunoResponsavel) TableName() string {
	return "aluno_responsaveis"
}

// AuditLog registra ações sensíveis da plataforma e das escolas.
// Diferente das entidades operacionais, ele é append-only por design:
// criamos eventos, não editamos nem removemos históricos.
type AuditLog struct {
	ID                uint `gorm:"primaryKey"`
	CreatedAt         time.Time
	ActorUserID       *uint    `gorm:"index"`
	ActorUser         *Usuario `gorm:"foreignKey:ActorUserID"`
	ActorPlatformRole string
	ActorSchoolRole   string
	ActorEscolaID     *uint   `gorm:"index"`
	ActorEscola       *Escola `gorm:"foreignKey:ActorEscolaID"`
	TargetEscolaID    *uint   `gorm:"index"`
	TargetEscola      *Escola `gorm:"foreignKey:TargetEscolaID"`
	Action            string  `gorm:"not null;index"`
	EntityType        string  `gorm:"not null;index"`
	EntityID          string
	Origin            string          `gorm:"not null"`
	BeforeJSON        json.RawMessage `gorm:"type:jsonb"`
	AfterJSON         json.RawMessage `gorm:"type:jsonb"`
	IP                string
	UserAgent         string
}
