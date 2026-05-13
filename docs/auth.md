# Auth & Login — Documentação Técnica

Última atualização: Abril 2026

---

## Visão Geral

O sistema de autenticação do Diário Kids usa **JWT (JSON Web Token)** com assinatura **HS256**. O cliente faz login com email e senha, recebe um Bearer Token e o envia no header `Authorization` em todas as requisições protegidas.

Não há refresh token por enquanto — o token tem validade de **24 horas**.

---

## Fluxo de Autenticação

```
Cliente
  │
  ├─ POST /api/login  { email, senha }
  │         │
  │     [Rate Limiter: 1 req/s, burst 5]
  │         │
  │     [AuthHandler] → valida body, valida DTO
  │         │
  │     [AuthService] → busca usuário, bcrypt, gera JWT
  │         │
  │     ◄── { token, email }
  │
  ├─ GET /api/rota-protegida
  │     Authorization: Bearer <token>
  │         │
  │     [Verificar middleware] → valida JWT, injeta usuário no contexto
  │         │
  │     [Handler protegido] → acessa r.Context() para dados do usuário
```

---

## Arquivos Relevantes

| Arquivo | Responsabilidade |
|---|---|
| `internal/handlers/auth_handler.go` | Camada HTTP: recebe requisição, valida input, retorna resposta |
| `internal/services/auth_service.go` | Regra de negócio: bcrypt, geração de JWT, sentinels de erro |
| `internal/middleware/auth_middleware.go` | Valida Bearer Token em rotas protegidas |
| `internal/middleware/rate_limiter.go` | Token Bucket por IP no endpoint de login |
| `internal/dto/auth_dto.go` | Structs de entrada (`LoginInput`) e saída (`LoginResponse`) |

---

## Endpoint de Login

```
POST /api/login
Content-Type: application/json

{
  "email": "usuario@escola.com",
  "senha": "minhasenha123"
}
```

**Respostas:**

| Status | Situação |
|---|---|
| `200 OK` | Login bem-sucedido. Body: `{ "token": "...", "email": "..." }` |
| `400 Bad Request` | JSON mal formatado ou campos desconhecidos no payload |
| `413 Request Entity Too Large` | Body acima de 1MB |
| `422 Unprocessable Entity` | Validação falhou (ex: email inválido, senha < 8 chars) |
| `429 Too Many Requests` | Rate limit atingido. Header `Retry-After: 60` incluso |
| `401 Unauthorized` | Credenciais inválidas (email não existe ou senha errada) |
| `500 Internal Server Error` | Erro interno (banco, JWT_SECRET ausente). Detalhes apenas nos logs |

---

## JWT

**Algoritmo:** HS256 (segredo simétrico via `JWT_SECRET` no `.env`)

**Payload (claims):**

```json
{
  "sub": "42",
  "email": "usuario@escola.com",
  "role": "professor",
  "iat": 1714300800,
  "exp": 1714387200
}
```

| Campo | Descrição |
|---|---|
| `sub` | ID do usuário no banco (como string) |
| `email` | Email do usuário |
| `role` | Papel: `pai`, `professor`, `coordenador`, `diretor`, `proprietario` |
| `iat` | Timestamp de criação |
| `exp` | Expira em 24h a partir da criação |

---

## Rotas Protegidas

Para acessar qualquer rota protegida, o cliente deve enviar o token no header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

O middleware `Verificar` (em `auth_middleware.go`) intercepta a requisição, valida o token e injeta os dados do usuário no contexto:

```go
// Dentro de um handler protegido:
usuario := r.Context().Value(middleware.ContextKeyUsuario).(middleware.UsuarioAutenticado)
// usuario.ID, usuario.Email, usuario.Role disponíveis aqui
```

---

## Segurança Implementada

### Timing Attack
O `AuthService` sempre executa o `bcrypt.CompareHashAndPassword`, mesmo quando o email não existe. Isso evita que um atacante descubra quais emails estão cadastrados medindo o tempo de resposta. Quando o usuário não existe, comparamos a senha com um `dummyHash` pré-computado — o tempo de resposta é ~200ms em ambos os casos.

### Algoritmo JWT
O middleware verifica explicitamente que o token usa `*jwt.SigningMethodHMAC`. Isso bloqueia o ataque `alg:none`, onde um atacante envia um token sem assinatura tentando se passar por autenticado.

### Rate Limiting
O endpoint `/api/login` tem um limitador por IP usando o algoritmo **Token Bucket**:
- **Taxa:** 1 requisição/segundo por IP
- **Burst:** 5 tentativas instantâneas permitidas antes de bloquear
- **Limpeza automática:** IPs inativos por mais de 3 minutos são removidos da memória

### Input Hardening
- Body limitado a **1MB** via `http.MaxBytesReader`
- `DisallowUnknownFields` rejeita campos não declarados no DTO
- Senha mínima de **8 caracteres** (alinhado com NIST SP 800-63B)

### Mensagens de erro
Nunca expõe `err.Error()` diretamente ao cliente. Erros internos são logados com `slog.Error` no servidor. Para o cliente, apenas mensagens genéricas classificadas por status HTTP.

---

## Variáveis de Ambiente

| Variável | Obrigatória | Descrição |
|---|---|---|
| `JWT_SECRET` | Sim | Segredo para assinar e verificar tokens. A API não sobe sem ele. |
| `DATABASE_URL` | Sim | URL de conexão com o banco Postgres (Supabase) |

---

## O que ainda NÃO foi implementado

- Refresh tokens
- Blacklist de tokens (ex: logout explícito)
- Complexidade de senha (maiúsculas, números, caracteres especiais)
- Bloqueio de conta após N tentativas falhas
- Autenticação via OAuth / provedores externos
