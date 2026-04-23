package database

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	/*
	   Precisamos dele lá no final do arquivo, na linha do 'AutoMigrate'.
	   Como vamos pedir pro GORM criar as tabelas baseadas nas Structs, temos que importar
	   o pacote onde essas Structs (Usuario, Sala, Aluno) foram escritas manualmente.
	*/
	"github.com/rafael-mingossi/diario-kids-api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Agora a função retorna o banco de dados e um possível erro.
func ConnectDB() (*gorm.DB, error) {
	// Pega a URL do arquivo .env ou do ambiente de deploy (Render, etc)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// O padrão de LOG (imprimir na tela) é o 'slog.Error()'.
		// O 'fmt.Errorf()' serve apenas para CRIAR um erro na memória do Go e
		// devolvê-lo (return) para quem chamou a função. Quem vai logar de verdade
		// é o main.go.
		return nil, fmt.Errorf("DATABASE_URL não encontrada nas variáveis de ambiente")
	}

	/*
		No Go, funções podem retornar múltiplos valores nativamente.
		O 'gorm.Open' sempre devolve (1) a Conexão e (2) um Erro.
		Poderia ser qualquer nome (ex: conexao, problema := gorm.Open), mas 'db' e 'err'
		  são a convenção oficial da comunidade.
	*/
	// Conecta ao Supabase usando o GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Erro ao conectar ao banco de dados: %v", err)
	}

	// 1. Extraímos o banco de baixo nível (sql.DB) de dentro do GORM
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("Erro ao extrair sql.DB do GORM: %w", err)
	}

	// 2. O PING: Garante que a ponte de rede até o Supabase existe de verdade
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao pingar o banco de dados: %w", err)
	}

	// 3. O POOL DE CONEXÕES: Protege o Supabase de sobrecarga
	sqlDB.SetMaxOpenConns(10)                 // Limite máximo de conexões simultâneas
	sqlDB.SetMaxIdleConns(5)                  // Conexões que ficam abertas em stand-by
	sqlDB.SetConnMaxLifetime(time.Minute * 5) // Renova as conexões a cada 5 minutos

	slog.Info("🐘 Conectado ao banco de dados Supabase com sucesso!")

	// 4. MIGRATIONS: O AutoMigrate olha os moldes do pacote 'models' e cria as tabelas no Postgres!
	slog.Info("🛠️  Verificando e sincronizando tabelas...")
	err = db.AutoMigrate(&models.Usuario{}, &models.Sala{}, &models.Aluno{})
	if err != nil {
		return nil, fmt.Errorf("erro durante as migrations: %w", err)
	}

	slog.Info("✅ Tabelas sincronizadas!")

	// Devolvemos a "chave" do banco pronta para uso
	return db, nil
}
