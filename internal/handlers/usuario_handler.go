package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rafael-mingossi/diario-kids-api/internal/models"
	"github.com/rafael-mingossi/diario-kids-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UsuarioHandler struct {
	repo *repository.UsuarioRepository
}

func NewUsuarioHandler(repo *repository.UsuarioRepository) *UsuarioHandler {
	return &UsuarioHandler{repo: repo}
}

// Struct (DTO) apenas para receber o JSON do frontend
type CadastroRequisicao struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
	Role  string `json:"role"` // "professor" ou "responsavel"
}

func (h *UsuarioHandler) CriarUsuario(w http.ResponseWriter, r *http.Request) {
	// 1. Lemos o JSON que veio do frontend (React/React Native)
	var req CadastroRequisicao

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 2. Criptografamos a senha (custa um pouco de processamento, o 14 é a "força" do hash)
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), 14)
	if err != nil {
		http.Error(w, "Erro ao criptografar a senha", http.StatusInternalServerError)
		return
	}

	// 3. Montamos o Usuário como o banco espera
	novoUsuario := models.Usuario{
		Nome:      req.Nome,
		Email:     req.Email,
		SenhaHash: string(senhaHash), // O hash é uma sequência de caracteres, então convertemos para string
		Role:      req.Role,
	}

	// 4. Mandamos o repositório salvar
	if err := h.repo.CriarUsuario(&novoUsuario); err != nil {
		http.Error(w, "Erro ao criar usuário (Email já existe?)", http.StatusConflict)
		return
	}

	// 5. Respondemos com sucesso!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "Usuário criado com sucesso!"})
}
