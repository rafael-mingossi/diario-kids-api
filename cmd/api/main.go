package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rafael-mingossi/diario-kids-api/internal/database"
	"github.com/rafael-mingossi/diario-kids-api/internal/handlers"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// =========================
	// 1. Variáveis de Ambiente
	// =========================

	//Carrega o arquivo .env se ele existir
	err := godotenv.Load()
	if err != nil {
		// Substituímos o fmt.Println pelo slog.Warn
		slog.Warn("⚠️  Aviso: Arquivo .env não encontrado, usando variáveis de sistema.")
	}

	// Conecta ao banco de dados (Supabase)
	db, err := database.ConnectDB()
	if err != nil {
		// Se o Ping falhar, ou a senha estiver errada, matamos o app aqui.
		slog.Error("Falha crítica ao iniciar banco de dados", "erro", err)
		os.Exit(1)
	}

	// === NOVIDADE: Injeção de Dependência ===
	usuarioRepo := repository.NewUsuarioRepository(db)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioRepo)
	// ========================================

	// Porta Dinâmica
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Se não achar a variável, usa a 8080 local
	}

	// =========================
	// 2. Setup do Servidor
	// =========================

	// Inicializando o roteador Chi
	r := chi.NewRouter()

	// Middlewares nativos muito úteis que já vêm no Chi:
	r.Use(middleware.Logger)    // Faz um "console.log" automático de cada requisição
	r.Use(middleware.Recoverer) // Se o app der um erro fatal (panic), ele não derruba o servidor inteiro

	// rota inicial
	r.Get("/api/status", handlers.StatusHandler)

	// === NOVIDADE: A nossa nova rota POST ===
	r.Post("/api/usuarios", usuarioHandler.CriarUsuario)
	// ========================================

	// Para termos controle sobre o desligamento, não podemos usar apenas
	// http.ListenAndServe. Precisamos criar uma "Instância" do servidor:
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// ====================
	// 3. Graceful Shutdown
	// ====================

	// Criamos um "Canal" (Channel). É assim que coisas rolando em paralelo se comunicam no Go.
	// Esse canal vai escutar sinais do Sistema Operacional.
	stop := make(chan os.Signal, 1)

	// Removido o os.Interrupt, mantendo apenas os sinais limpos do sistema
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// A palavra 'go' cria uma Goroutine. É como uma thread rodando em background.
	// Ou seja, o servidor vai subir rodando "de lado", sem travar o código principal.
	go func() {
		slog.Info("API DK iniciada", "porta", port)
		// ErrServerClosed é o erro normal de quando pedimos pro servidor desligar.
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Erro ao iniciar o servidor", "erro", err)
			os.Exit(1)
		}
	}()

	// O código principal chega aqui e "trava".
	// O '<-stop' diz: "Fique esperando até chegar alguma coisa no canal 'stop'."
	<-stop

	// Se chegou aqui, é porque você apertou Ctrl+C ou o Docker mandou parar.
	slog.Info("🛑 Sinal recebido. Desligando a API graciosamente...")

	// Criamos um contexto com tempo limite de 5 segundos para o desligamento.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Pedimos para o servidor desligar esperando o tempo do contexto acima.
	// Ele termina os requests pendentes e fecha as conexões.
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Erro durante o Graceful Shutdown", "erro", err)
	}
	slog.Info("API DK desligada com sucesso! 👋")
}
