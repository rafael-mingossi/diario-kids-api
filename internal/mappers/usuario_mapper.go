// Package mappers é responsável exclusivamente por transformações de dados entre camadas.
// Regra de ouro: nenhuma função aqui deve conter lógica de negócio (sem bcrypt, sem validações,
// sem chamadas ao banco). Apenas copia e reorganiza campos entre tipos.
package mappers

import (
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/models"
)

// CriarInputToModel converte um DTO de entrada em um Model pronto para persistência.
//
// Por que o senhaHash é passado separado?
// A senha em texto puro (input.Senha) é criptografada pelo Service antes de chegar aqui.
// O Mapper não tem responsabilidade de criptografar — ele apenas monta a struct.
// Isso mantém o Mapper puro e testável sem depender do bcrypt.
func CriarInputToModel(input dto.CriarUsuarioInput, senhaHash string) models.Usuario {
	return models.Usuario{
		Nome:      input.Nome,
		Email:     input.Email,
		SenhaHash: senhaHash, // Já chegou criptografado do Service
		Role:      input.Role,
	}
}

// ModelToUsuarioResponse converte um Model de banco em um DTO de resposta seguro.
//
// Este é o ponto de segurança da saída: ao mapear explicitamente campo a campo,
// garantimos que campos sensíveis como SenhaHash NUNCA vazam para o cliente,
// mesmo que o model evolua e ganhe novos campos internos no futuro.
func ModelToUsuarioResponse(u models.Usuario) dto.UsuarioResponse {
	return dto.UsuarioResponse{
		ID:    u.ID,   // ID gerado pelo banco após o INSERT
		Nome:  u.Nome,
		Email: u.Email,
		Role:  u.Role,
		// SenhaHash é intencionalmente omitido aqui
	}
}

// UsuarioModelsToResponseList converte uma slice de Models em uma slice de DTOs de resposta.
//
// Usado em endpoints de listagem (ex: GET /api/usuarios).
// Ao centralizar a conversão aqui, garantimos consistência: todas as listagens
// aplicam as mesmas regras de omissão de campos sensíveis que o ModelToUsuarioResponse acima.
// Se a regra de saída mudar, muda em um único lugar.
func UsuarioModelsToResponseList(usuarios []models.Usuario) []dto.UsuarioResponse {
	// Pré-alocamos a slice com o tamanho exato para evitar realocações de memória
	// durante o append — uma boa prática quando o tamanho é conhecido.
	resposta := make([]dto.UsuarioResponse, len(usuarios))

	for i, u := range usuarios {
		// Reutilizamos o mapper de item único para não duplicar a lógica de mapeamento
		resposta[i] = ModelToUsuarioResponse(u)
	}

	return resposta
}
