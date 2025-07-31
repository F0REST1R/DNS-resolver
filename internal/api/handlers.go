package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type AddFQDNRequest struct {
	FQDN string `json:"fqdn" validate:"required,fqdn"`
}

func (h *Handler) AddFQDN(c echo.Context) error {
	var req AddFQDNRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	ips, err := h.resolver.Resolve(ctx, req.FQDN)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "DNS resolution failed")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"fqdn": req.FQDN,
		"ips":  ips,
	})
}

func (h *Handler) GetFQDNsByIP(c echo.Context) error {
	ip := c.QueryParam("ip")
	if ip == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ip parameter is required")
	}

	ctx := c.Request().Context()
	fqdns, err := h.resolver.GetFQDNsByIP(ctx, ip)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "db error")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"ip":    ip,
		"fqdns": fqdns,
	})
}

func (h *Handler) GetIPsByFQDN(c echo.Context) error {
	fqdn := c.QueryParam("fqdn")
	if fqdn == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "fqdn parameter is required")
	}

	ctx := c.Request().Context()
	ips, err := h.resolver.GetIPsByFQDN(ctx, fqdn)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "db error")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"fqdn": fqdn,
		"ips":    ips,
	})
}
