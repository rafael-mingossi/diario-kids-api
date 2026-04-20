package handlers

import (
	"net/http"
)

// StatusHandler é o que chamamos de "Handler".
// Ele segue a assinatura padrão do Go: (ResponseWriter, *Request)
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "A API do DK está funcionando!"}`))
}
