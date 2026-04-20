package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rafael-mingossi/diario-kids-api/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Inicializando o roteador Chi
	r := chi.NewRouter()

	// Middlewares nativos muito úteis que já vêm no Chi:
	r.Use(middleware.Logger)    // Faz um "console.log" automático de cada requisição
	r.Use(middleware.Recoverer) // Se o app der um erro fatal (panic), ele não derruba o servidor inteiro

	// rota inicial
	r.Get("/api/status", handlers.StatusHandler)

	fmt.Println("A API do DK rodando na porta 8080...")

	// Subindo o servidor com o nosso novo roteador 'r'
	// O log.Fatal faz duas coisas:
	// 1. Imprime a mensagem de erro com a data e hora.
	// 2. Encerra o programa (código de saída 1) imediatamente.
	log.Fatal(http.ListenAndServe(":8080", r))
}
