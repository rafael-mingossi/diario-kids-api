package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Criamos uma variável global para guardar a conexão e reaproveitá-la
var DB *gorm.DB

func ConnectDB() {
	// Pega a URL do arquivo .env ou do ambiente de deploy (Render, etc)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL não encontrada. Verifique seu arquivo .env ou variáveis de ambiente.")
	}

	// Conecta ao Supabase usando o GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	fmt.Println("🐘 Conectado ao banco de dados Supabase com sucesso!")

	// Guarda a conexão na variável global para usarmos nos nossos Repositórios depois
	DB = db

}
