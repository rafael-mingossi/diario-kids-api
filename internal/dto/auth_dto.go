package dto

// Receber do Frontend
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Senha    string `json:"senha" validate:"required,min=8"`
	EscolaID *uint  `json:"escola_id,omitempty"`
}

// Devolver para o Frontend
type LoginResponse struct {
	Token        string `json:"token"`
	Email        string `json:"email"`
	Role         string `json:"role,omitempty"`
	EscolaID     *uint  `json:"escola_id,omitempty"`
	PlatformRole string `json:"platform_role,omitempty"`
}
