package models

import (
	"time"

	"gorm.io/gorm"
)

type DNSRecord struct {
	gorm.Model
	FQDN       string    `gorm:"not null;index"`
	IP         string    `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdateAt  time.Time `gorm:"autoUpdateTime"`
}
