package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Inicializando o roteador Chi
	r := chi.NewRouter()

	// Middlewares nativos muito úteis que já vêm no Chi:
	r.Use(middleware.Logger)    // Faz um "console.log" automático de cada requisição
	r.Use(middleware.Recoverer) // Se o app der um erro fatal (panic), ele não derruba o servidor inteiro

	// Nossa rota inicial
	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"mensagem": "A API do DK está funcionando!"}`))
	})

	fmt.Println("A API do DK rodando na porta 8080...")

	// Subindo o servidor com o nosso novo roteador 'r'
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println("Erro ao iniciar o server: ", err)
	}
}
