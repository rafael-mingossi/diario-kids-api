package dto

// CriarEscolaInput recebe tanto matriz quanto filial.
// A diferença entre elas é decidida pelas regras do service:
// - matriz: IsMatriz=true e MatrizID=nil
// - filial: IsMatriz=false e MatrizID apontando para a matriz dona
type CriarEscolaInput struct {
	ClienteID    uint   `json:"cliente_id" validate:"required"`
	CNPJ         string `json:"cnpj" validate:"required"`
	RazaoSocial  string `json:"razao_social" validate:"required,min=3"`
	NomeFantasia string `json:"nome_fantasia" validate:"required,min=3"`
	IsMatriz     bool   `json:"is_matriz"`
	MatrizID     *uint  `json:"matriz_id,omitempty"`
}

type EscolaResponse struct {
	ID           uint   `json:"id"`
	ClienteID    uint   `json:"cliente_id"`
	CNPJ         string `json:"cnpj"`
	RazaoSocial  string `json:"razao_social"`
	NomeFantasia string `json:"nome_fantasia"`
	IsMatriz     bool   `json:"is_matriz"`
	MatrizID     *uint  `json:"matriz_id,omitempty"`
}
