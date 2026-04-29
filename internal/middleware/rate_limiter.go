package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	// golang.org/x/time/rate é a biblioteca oficial do time da Go para rate limiting.
	// Ela implementa o algoritmo "Token Bucket" (Balde de Fichas):
	//   - Imagine um balde com capacidade máxima de N fichas.
	//   - Cada requisição consome 1 ficha.
	//   - O balde se reabastece automaticamente a uma taxa constante (ex: 1 ficha/segundo).
	//   - Quando o balde esvazia, as requisições são recusadas até ele reabastecer.
	// É o mesmo algoritmo usado por AWS API Gateway, Cloudflare, etc.
	"golang.org/x/time/rate"
)

// visitante representa o estado do rate limiter de um IP específico.
// Guardamos o limiter em si e a última vez que esse IP fez uma requisição.
// O campo 'ultimaVez' serve para limparmos visitantes inativos da memória
// e evitar um vazamento de memória com o tempo.
type visitante struct {
	limiter   *rate.Limiter
	ultimaVez time.Time
}

// LimitadorPorIP é a struct que gerencia um map de visitantes, um por IP.
// O sync.Mutex garante que múltiplas goroutines (requisições simultâneas)
// não corrompam o mapa ao ler e escrever ao mesmo tempo.
//
// Analogia JS: é como um Map<string, { limiter, lastSeen }> protegido por um Lock,
// já que Go é concorrente por natureza — diferente do Node.js que é single-thread.
type LimitadorPorIP struct {
	visitantes map[string]*visitante
	mu         sync.Mutex
	// r: taxa de reabastecimento do balde (fichas por segundo)
	// b: capacidade máxima do balde (burst — pico máximo tolerado)
	r rate.Limit
	b int
}

// NovoLimitadorPorIP cria um LimitadorPorIP e já inicia uma goroutine de limpeza.
// Parâmetros:
//   - r: quantas requisições por segundo são permitidas (ex: 5 = 5 req/s)
//   - b: burst máximo (ex: 10 = pode ter até 10 requisições instantâneas antes de bloquear)
func NovoLimitadorPorIP(r rate.Limit, b int) *LimitadorPorIP {
	l := &LimitadorPorIP{
		visitantes: make(map[string]*visitante),
		r:          r,
		b:          b,
	}

	// Iniciamos uma goroutine que limpa IPs inativos a cada minuto.
	// Sem isso, cada IP que já fez uma requisição ficaria na memória para sempre
	// — um vazamento de memória lento mas garantido em produção.
	go l.limparInativos()

	return l
}

// obterLimiter retorna o limitador para um IP específico, criando um se não existir.
// É thread-safe graças ao mutex.
func (l *LimitadorPorIP) obterLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock() // defer garante que o unlock acontece ao sair da função, sempre

	v, existe := l.visitantes[ip]
	if !existe {
		// Primeira vez que vemos este IP — criamos um novo balde para ele
		limiter := rate.NewLimiter(l.r, l.b)
		l.visitantes[ip] = &visitante{limiter, time.Now()}
		return limiter
	}

	// IP já conhecido — atualizamos a última vez que foi visto e devolvemos o balde
	v.ultimaVez = time.Now()
	return v.limiter
}

// limparInativos remove IPs que não fazem requisições há mais de 3 minutos.
// Roda em background como uma goroutine independente.
func (l *LimitadorPorIP) limparInativos() {
	// time.Tick retorna um canal que recebe um valor a cada intervalo.
	// Analogia JS: é como um setInterval(() => { ... }, 60000) que roda para sempre.
	for range time.Tick(time.Minute) {
		l.mu.Lock()
		for ip, v := range l.visitantes {
			if time.Since(v.ultimaVez) > 3*time.Minute {
				delete(l.visitantes, ip)
			}
		}
		l.mu.Unlock()
	}
}

// extrairIP obtém o IP real do cliente, considerando proxies e load balancers.
// Em produção (ex: Render, Railway), o IP real vem no header X-Forwarded-For,
// não diretamente do RemoteAddr (que seria o IP do proxy).
func extrairIP(r *http.Request) string {
	// X-Forwarded-For pode ter múltiplos IPs separados por vírgula (ex: "client, proxy1, proxy2").
	// O IP do cliente real é sempre o primeiro da lista.
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	// Sem proxy: usamos o RemoteAddr direto.
	// RemoteAddr tem o formato "IP:porta", então removemos a porta.
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// LimitarRequisicoes retorna um middleware HTTP que aplica rate limiting por IP.
// Uso: aplique diretamente na rota que quer proteger (ex: login).
//
// Exemplo de uso no main.go:
//
//	loginLimiter := middleware.NovoLimitadorPorIP(5, 10) // 5 req/s, burst de 10
//	r.With(loginLimiter.LimitarRequisicoes()).Post("/api/login", authHandler.Login)
func (l *LimitadorPorIP) LimitarRequisicoes() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extrairIP(r)
			limiter := l.obterLimiter(ip)

			// limiter.Allow() tenta consumir 1 ficha do balde deste IP.
			// Se o balde estiver vazio (IP fez requisições demais), retorna false.
			if !limiter.Allow() {
				slog.Warn("rate limit atingido",
					"ip", ip,
					"path", r.URL.Path,
				)
				// 429 Too Many Requests é o status HTTP padrão para rate limiting.
				// O header Retry-After informa ao cliente quantos segundos esperar.
				// Usamos 60s como valor fixo conservador.
				w.Header().Set("Retry-After", "60")
				http.Error(w, "muitas tentativas. Tente novamente em 1 minuto.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
