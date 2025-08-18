package database

import (
	"discord-bot-go/internal/domain"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "discordpass"),
		getEnv("DB_NAME", "discordbot"),
		getEnv("DB_PORT", "5432"))

	// configurando gorm
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// conectando no DB
	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), config)
		if err == nil {
			break
		}

		log.Printf("Falha ao conectar no DB (tentativa %d/10): %v", i+1, err)
		time.Sleep(time.Duration(i) * time.Second)
	}

	if err != nil {
		log.Fatal("Falha ao conectar no DB:", err)
	}

	log.Println("Connected to database successfully!")

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// migração
	err = DB.AutoMigrate(&domain.Anger{}, &domain.Birthday{}, &domain.Server{}, &domain.LastMatch{}, &domain.Player{})
	if err != nil {
		log.Fatal("Falha migrando DB:", err)
	}

	log.Println("Migração completa")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
