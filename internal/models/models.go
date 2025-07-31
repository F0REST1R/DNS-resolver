package models

import (
	"context"
	"net"
	"time"

	"gorm.io/gorm"
)

type DNSRecord struct {
	ID        uint      `gorm:"primarykey"`
	FQDN      string    `gorm:"not null;index"`
	IP        string    `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

type DNSResolver struct {
	DB             *gorm.DB
	Resolver       *net.Resolver
	UpdateInterval time.Duration
}

type Repository interface {
	AddOrUpdate(ctx context.Context, fqdn, ip string) error
	GetIPsByFQDN(ctx context.Context, fqdn string) ([]string, error)
	GetFQDNsByIP(ctx context.Context, ip string) ([]string, error)
	GetAllFQDNs(ctx context.Context) ([]string, error)
}
