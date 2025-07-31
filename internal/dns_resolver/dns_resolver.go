package dnsresolver

import (
	"context"
	repo "dns-resolver/internal/repository"
	"net"
)

type Resolver struct {
	repo repo.Repository
}

func NewResolver(repo repo.Repository) *Resolver {
	return &Resolver{repo: repo}
}

func (r *Resolver) Resolve(ctx context.Context, fqdn string) ([]string, error) {
	ips, err := net.LookupIP(fqdn)
	if err != nil {
		return nil, err
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
		r.repo.AddOrUpdate(ctx, fqdn, ip.String())
	}

	return ipStrings, nil
}

// GetFQDNsByIP проксирует запрос в репозиторий
func (r *Resolver) GetFQDNsByIP(ctx context.Context, ip string) ([]string, error) {
	return r.repo.GetFQDNsByIP(ctx, ip)
}

// Аналогично для других методов
func (r *Resolver) GetIPsByFQDN(ctx context.Context, fqdn string) ([]string, error) {
	return r.repo.GetIPsByFQDN(ctx, fqdn)
}

func (r *Resolver) AddOrUpdate(ctx context.Context, fqdn, ip string) error {
	return r.repo.AddOrUpdate(ctx, fqdn, ip)
}
