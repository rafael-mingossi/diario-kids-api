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
	// Importamos nosso middleware com um alias 'authmiddleware' para não colidir
	// com o pacote de middleware do Chi que também se chama 'middleware'.
	authmiddleware "github.com/rafael-mingossi/diario-kids-api/internal/middleware"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
	"github.com/rafael-mingossi/diario-kids-api/internal/services"

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

	// Validação de segurança: o JWT_SECRET é obrigatório para a API funcionar.
	// Se não estiver definido, matamos o processo na inicialização — é muito melhor
	// descobrir isso agora do que servir tokens inválidos em produção.
	// Analogia JS: como um `if (!process.env.JWT_SECRET) throw new Error(...)` no topo do app.
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET não encontrado. Defina-o no .env ou nas variáveis de ambiente.")
		os.Exit(1)
	}

	// Conecta ao banco de dados (Supabase)
	db, err := database.ConnectDB()
	if err != nil {
		// Se o Ping falhar, ou a senha estiver errada, matamos o app aqui.
		slog.Error("Falha crítica ao iniciar banco de dados", "erro", err)
		os.Exit(1)
	}

	// ==========================================
	// INJEÇÃO DE DEPENDÊNCIAS
	// ==========================================

	// 1. Repositórios (Acesso a dados)
	usuarioRepo := repository.NewUsuarioRepository(db)
	salaRepo := repository.NewSalaRepository(db)

	// 2. Serviços (Regras de negócio)
	usuarioService := services.NewUsuarioService(usuarioRepo)
	authService := services.NewAuthService(usuarioRepo)
	salaService := services.NewSalaService(salaRepo)

	// 3. Handlers (Recepção HTTP)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioService)
	authHandler := handlers.NewAuthHandler(authService)
	salaHandler := handlers.NewSalaHandler(salaService)

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

	// ==========================================
	// ROTAS PÚBLICAS — não precisam de token JWT
	// ==========================================

	// Health check — sempre acessível
	r.Get("/api/status", handlers.StatusHandler)

	// Registro de novo usuário — público para permitir o cadastro inicial
	r.Post("/api/usuarios", usuarioHandler.CriarUsuario)

	// Login com rate limiting por IP.
	//
	// NovoLimitadorPorIP(1, 5) significa:
	//   - 1 = taxa de reabastecimento: 1 ficha por segundo por IP
	//   - 5 = burst: um IP pode fazer até 5 tentativas instantâneas antes de ser bloqueado
	//
	// Valores conservadores adequados para login:
	//   - Um humano que errou a senha 3 vezes ainda passa (burst de 5 cobre isso).
	//   - Após esgotar o burst, só consegue 1 nova tentativa por segundo.
	//   - Scripts de brute force são bloqueados após a 5ª tentativa.
	//
	// r.With() aplica o middleware APENAS para esta rota específica.
	// Analogia JS: router.post('/api/login', rateLimitMiddleware, authHandler.login)
	loginLimiter := authmiddleware.NovoLimitadorPorIP(1, 5)
	r.With(loginLimiter.LimitarRequisicoes()).Post("/api/login", authHandler.Login)

	// ==========================================
	// ROTAS PROTEGIDAS — exigem token JWT válido
	// ==========================================
	// r.Group cria um sub-roteador isolado onde aplicamos o middleware apenas
	// para as rotas dentro do bloco. Rotas fora do grupo não são afetadas.
	//
	// Analogia JS: é como criar um router separado no Express e fazer
	// router.use(authenticateToken) antes de registrar as rotas protegidas,
	// depois montar com app.use('/api', router).
	r.Group(func(r chi.Router) {
		// Aplicamos o middleware de verificação JWT a todas as rotas deste grupo.
		// Passamos o secret lido do ambiente — injeção de dependência, não global.
		r.Use(authmiddleware.Verificar(jwtSecret))

		// Sala — apenas usuários autenticados podem criar salas
		r.Post("/api/salas", salaHandler.CriarSala)
	})

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
