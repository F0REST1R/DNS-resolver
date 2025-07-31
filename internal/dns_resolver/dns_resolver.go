package dnsresolver

import (
	"context"
	repo "dns-resolver/internal/repository"
	"log"
	"net"
	"os"
	"time"
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


func (r *Resolver) DNSUpdater(ctx context.Context, interval time.Duration) {
	logger := log.New(os.Stdout, "DNS_UPDATER: ", log.LstdFlags|log.Lshortfile)
    logger.Printf("Starting DNS updater with interval %v", interval)

    ticker := time.NewTicker(interval)
    defer func() {
        ticker.Stop()
        logger.Println("DNS updater stopped")
    }()

    for {
        select {
        case <-ticker.C:
            startTime := time.Now()
            logger.Println("Starting DNS records update cycle...")

            fqdns, err := r.GetAllFQDNs(ctx)
            if err != nil {
                logger.Printf("Failed to get FQDNs: %v", err)
                continue
            }

            logger.Printf("Found %d FQDNs to update", len(fqdns))
            
            successCount := 0
            for _, fqdn := range fqdns {
                select {
                case <-ctx.Done():
                    logger.Println("Update cycle interrupted by context")
                    return
                default:
                    ips, err := r.Resolve(ctx, fqdn)
                    if err != nil {
                        logger.Printf("Failed to resolve %s: %v", fqdn, err)
                        continue
                    }
                    successCount++
                    logger.Printf("Updated %s -> %v", fqdn, ips)
                }
            }

            logger.Printf("Update cycle completed. Success: %d/%d, Duration: %v", 
                successCount, len(fqdns), time.Since(startTime))

        case <-ctx.Done():
            logger.Println("Shutting down DNS updater by context signal")
            return
        }
    }
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

func (r *Resolver) GetAllFQDNs(ctx context.Context) ([]string, error) {
	return r.repo.GetAllFQDNs(ctx)
}