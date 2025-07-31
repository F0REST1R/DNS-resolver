package api

import (
	dnsresolver "dns-resolver/internal/dns_resolver"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	resolver *dnsresolver.Resolver
}

func NewHandler(resolver *dnsresolver.Resolver) *Handler {
	return &Handler{resolver: resolver}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.POST("/api/fqdns", h.AddFQDN)
	e.GET("/api/fqdns", h.GetFQDNsByIP)
	e.GET("/api/ips", h.GetIPsByFQDN)
}