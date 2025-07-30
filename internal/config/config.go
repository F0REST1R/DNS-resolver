package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func findProjectRoot(startDir string) (string, error) {
	markers := []string{
		"go.mod",
		"frontend",
		"auth-service",
	}

	currentDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(currentDir, marker)); err == nil {
				return currentDir, nil
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	return "", fmt.Errorf("Не удалось найти корень проекта")
}

func ProdDB() (*gorm.DB, error) {
	projectRoot, _ := findProjectRoot(".")
	envPath := filepath.Join(projectRoot, ".env")
	err := godotenv.Load(envPath)
	if err != nil{
		log.Fatal("Ошибка загрузки файла .env")
	}	

	dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s",
        os.Getenv("HOST"),
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_USER"),
        os.Getenv("DB_PORT"),
    )
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func TestDBcon() (*gorm.DB, error) {
	projectRoot, _ := findProjectRoot(".")
	envPath := filepath.Join(projectRoot, ".env")
	err := godotenv.Load(envPath)
	if err != nil{
		log.Fatal("Ошибка загрузки файла .env")
	}

	dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=%s port=5432 sslmode=disable", os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_USER"))
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}