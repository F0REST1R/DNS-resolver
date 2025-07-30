package repository

import (
	"context"
	"dns-resolver/internal/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func NewDB(db *gorm.DB) *DB{
	return &DB{db: db}
}

func NewPostgresDB(host, port, user, password, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", 
		host, user, password, dbname, port, 
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return db, nil
}

func (d *DB) GetFQDNsByIP(ctx context.Context, ip string) ([]string, error) {
	var records []models.DNSRecord
	err := d.db.WithContext(ctx).Where("ip = ?", ip).Find(&records).Error
	if err != nil {
		return nil, err
	}

	fqdns := make([]string, len(records))
	for i, record := range records {
		fqdns[i] = record.FQDN
	}

	return fqdns, nil
}

func (d *DB) GetIPsByFQDN(ctx context.Context, fqdn string) ([]string, error) {
	var records []models.DNSRecord
	err := d.db.WithContext(ctx).Where("fqdn = ?", fqdn).Find(&records).Error
	if err != nil{
		return nil, err
	}

	ips := make([]string, len(records))
	for i, record := range records {
		ips[i] = record.IP
	}

	return ips, nil
}

// AddOrUpdateRecord добавляет или обновляет запись
func (d *DB) AddOrUpdate(ctx context.Context, fqdn, ip string) error {
	return d.db.WithContext(ctx).Where(models.DNSRecord{FQDN: fqdn, IP: ip}).FirstOrCreate(&models.DNSRecord{FQDN: fqdn, IP: ip}).Error
}

