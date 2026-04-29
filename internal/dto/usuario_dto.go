package dto

// O que esperamos receber do Frontend (React Native/Next.js)
type CriarUsuarioInput struct {
	Nome  string `json:"nome" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Senha string `json:"senha" validate:"required,min=8"`
	Role  string `json:"role" validate:"required,oneof=pai professor coordenador diretor proprietario"`
}

// O que vamos devolver para o Frontend (Escondendo dados sensíveis)
type UsuarioResponse struct {
	ID    uint   `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
