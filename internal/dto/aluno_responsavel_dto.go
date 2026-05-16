package dto

type VincularAlunoResponsavelInput struct {
	AlunoID            uint   `json:"aluno_id" validate:"required"`
	UsuarioID          uint   `json:"usuario_id" validate:"required"`
	Parentesco         string `json:"parentesco" validate:"required,oneof=pai mae avo ava tio tia guardiao responsavel_legal outro"`
	ResponsavelLegal   bool   `json:"responsavel_legal"`
	RecebeNotificacoes bool   `json:"recebe_notificacoes"`
	ContatoEmergencia  bool   `json:"contato_emergencia"`
	AutorizadoBusca    bool   `json:"autorizado_busca"`
	Observacao         string `json:"observacao,omitempty"`
	Ativo              *bool  `json:"ativo,omitempty"`
}

type AlunoResponsavelResponse struct {
	ID                 uint   `json:"id"`
	AlunoID            uint   `json:"aluno_id"`
	UsuarioID          uint   `json:"usuario_id"`
	EscolaID           uint   `json:"escola_id"`
	Parentesco         string `json:"parentesco"`
	ResponsavelLegal   bool   `json:"responsavel_legal"`
	RecebeNotificacoes bool   `json:"recebe_notificacoes"`
	ContatoEmergencia  bool   `json:"contato_emergencia"`
	AutorizadoBusca    bool   `json:"autorizado_busca"`
	Observacao         string `json:"observacao,omitempty"`
	Ativo              bool   `json:"ativo"`
}
