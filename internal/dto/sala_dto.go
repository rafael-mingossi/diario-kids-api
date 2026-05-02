package dto

// DTO é contrato de entrada/saída HTTP, não espelho 1:1 da tabela.
// DTO expõe só o contrato necessário

// receber do Frontend
type CriarSalaInput struct {
	Nome        string `json:"nome" validate:"required,min=3"`
	Numero      string `json:"numero" validate:"required"`
	ProfessorID *uint  `json:"professor_id,omitempty"`
}

// devolver para o Frontend
type SalaResponse struct {
	ID          uint   `json:"id"`
	Nome        string `json:"nome"`
	Numero      string `json:"numero"`
	ProfessorID *uint  `json:"professor_id,omitempty"`
}
