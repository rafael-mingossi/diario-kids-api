package dto

type CriarClienteInput struct {
	Nome            string `json:"nome" validate:"required,min=3"`
	Documento       string `json:"documento" validate:"required,min=11"`
	EmailFinanceiro string `json:"email_financeiro,omitempty" validate:"omitempty,email"`
	Telefone        string `json:"telefone,omitempty"`
	Status          string `json:"status,omitempty" validate:"omitempty,oneof=trial ativo suspenso cancelado"`
	Plano           string `json:"plano,omitempty" validate:"omitempty,oneof=basico premium enterprise demo"`
}

type ClienteResponse struct {
	ID              uint   `json:"id"`
	Nome            string `json:"nome"`
	Documento       string `json:"documento"`
	EmailFinanceiro string `json:"email_financeiro,omitempty"`
	Telefone        string `json:"telefone,omitempty"`
	Status          string `json:"status"`
	Plano           string `json:"plano"`
}
