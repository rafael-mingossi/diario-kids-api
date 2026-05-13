package dto

// SetupInicialInput cria o primeiro operador global da plataforma.
// Esse fluxo é de uso único e só funciona enquanto o banco ainda está vazio.
type SetupInicialInput struct {
	Nome  string `json:"nome" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Senha string `json:"senha" validate:"required,min=8"`
}

type SetupInicialResponse struct {
	UsuarioID    uint   `json:"usuario_id"`
	Email        string `json:"email"`
	PlatformRole string `json:"platform_role"`
	Token        string `json:"token"`
}
