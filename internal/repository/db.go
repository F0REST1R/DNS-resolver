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

func ProdDB() (*gorm.DB, error) {
	dsn := "host=postgres user=postgres password=dbdns dbname=DNS_DB port=5432 sslmode=require sslmode=disable"
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func DBForTest() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=dbdns dbname=DNS_DB port=5432 sslmode=require sslmode=disable"
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

func (d *DB) GetAllFQDNs(ctx context.Context) ([]string, error) {
	var fqdns []string
	err := d.db.WithContext(ctx).Model(&models.DNSRecord{}).Distinct("fqdn").Pluck("fqdn", &fqdns).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get FQDNs: %w", err)
	}

	return fqdns, nil
}