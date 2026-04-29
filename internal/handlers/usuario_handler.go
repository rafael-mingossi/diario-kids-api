package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rafael-mingossi/diario-kids-api/internal/dto"
	"github.com/rafael-mingossi/diario-kids-api/internal/services"
)

type UsuarioHandler struct {
	// NOVIDADE: Agora usamos a Interface (sem o asterisco *)
	service  services.UsuarioService
	validate *validator.Validate
}

//Handler has no interface and no private struct because nothing depends on it.

// Recebe o Service e inicializa o Validador
// Retorna um endereço de memória que aponta para um UsuarioHandler
// O * no retorno significa "Aviso: quem chamar essa função vai receber um endereço, não a struct inteira".
// Se a função retornasse apenas UsuarioHandler (sem *), ela criaria o objeto e entregaria uma cópia pesada para quem a chamou.
// Retornando o ponteiro, você passa o objeto recém-criado adiante de forma super leve.
func NewUsuarioHandler(service services.UsuarioService) *UsuarioHandler {
	// O & comercial significa "Crie isso na memória e me dê o endereço de onde ficou"
	return &UsuarioHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *UsuarioHandler) CriarUsuario(w http.ResponseWriter, r *http.Request) {
	var input dto.CriarUsuarioInput

	// 1. Limita o tamanho do body para prevenir "body bomb" (DoS por payload gigante).
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB

	// 2. Lê o JSON (O Garçom anota o pedido).
	// DisallowUnknownFields rejeita campos extras não declarados no DTO.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "corpo da requisição muito grande", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "JSON mal formatado", http.StatusBadRequest)
		return
	}

	// 2. Valida o input automaticamente usando as tags do DTO
	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "Dados de entrada inválidos (verifique email, tamanho da senha, etc)", http.StatusUnprocessableEntity)
		return
	}

	// 3. Manda para a Cozinha (Service) fazer o trabalho pesado
	resposta, err := h.service.CriarUsuario(input)
	if err != nil {
		// Feedback 2: Tratamos o erro específico de conflito (HTTP 409)
		if errors.Is(err, services.ErrEmailEmUso) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		// Feedback 1: Segurança (OWASP) - Escondemos o erro real do usuário, logamos no terminal
		slog.Error("Erro interno ao criar usuário", "detalhe", err)
		http.Error(w, "Erro interno no servidor. Tente novamente mais tarde.", http.StatusInternalServerError)
		return
	}

	// 4. Serializamos a resposta ANTES de escrever qualquer header.
	// Mesmo padrão do auth_handler: se o Marshal falhar, ainda podemos retornar 500.
	// Se chamarmos WriteHeader(201) primeiro e o Marshal falhar depois, o cliente
	// recebe um 201 com corpo vazio — inconsistência difícil de depurar.
	corpo, err := json.Marshal(resposta)
	if err != nil {
		slog.Error("erro ao serializar resposta de criação de usuário", "detalhe", err)
		http.Error(w, "erro interno no servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(corpo) //nolint:errcheck // Erro de escrita de rede após WriteHeader não é acionável
}
