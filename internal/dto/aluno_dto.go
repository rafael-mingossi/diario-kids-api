package dto

// receber do Frontend
type CriarAlunoInput struct {
	EscolaID uint   `json:"escola_id" validate:"required"`
	Nome string `json:"nome" validate:"required,min=3"`
	// O frontend envia a data como texto no formato brasileiro.
	// Exemplo: "25/12/2020". O parse para time.Time acontece no service.
	DataNascimento string `json:"data_nascimento" validate:"required,datetime=02/01/2006"`
	SalaID         *uint  `json:"sala_id,omitempty"`
}

// devolver para o Frontend

type AlunoResponse struct {
	ID             uint   `json:"id"`
	EscolaID       *uint  `json:"escola_id,omitempty"`
	Nome           string `json:"nome"`
	DataNascimento string `json:"data_nascimento"`
	SalaID         *uint  `json:"sala_id,omitempty"`
	Matricula      string `json:"matricula"`
}
