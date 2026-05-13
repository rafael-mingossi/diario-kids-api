package handlers

import (
	"encoding/json"
	"net/http"
)

// errorResponse padroniza os erros HTTP da API em JSON.
// Isso melhora debug no frontend/Postman/curl+jq, porque a resposta de erro
// deixa de ser texto puro e passa a ter estrutura previsível.
type errorResponse struct {
	Error string `json:"error"`
	Status int   `json:"status"`
}

// writeJSONError serializa a resposta de erro ANTES de escrever o status.
// É o mesmo padrão seguro que já usamos nas respostas de sucesso.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	corpo, err := json.Marshal(errorResponse{
		Error:  message,
		Status: status,
	})
	if err != nil {
		// Último fallback: se até o JSON do erro falhar, devolvemos status puro.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(corpo) //nolint:errcheck // Erro de escrita após WriteHeader não é acionável
}

// WriteJSONAuthError é exportado para o middleware porque middlewares vivem em outro pacote.
// Mantemos a implementação centralizada aqui para toda a API usar o mesmo formato de erro.
func WriteJSONAuthError(w http.ResponseWriter, status int, message string) {
	writeJSONError(w, status, message)
}