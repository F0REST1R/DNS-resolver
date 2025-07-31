package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ProdDB() (*gorm.DB, error) {
	dsn := "host=postgres user=postgres password=dbdns dbname=DNS_DB port=5432"
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func TestDBcon() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=dbdns dbname=DNS_DB port=5432 sslmode=disable"
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}