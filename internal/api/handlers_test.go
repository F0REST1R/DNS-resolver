package api

import (
	"context"
	dnsresolver "dns-resolver/internal/dns_resolver"
	v "dns-resolver/internal/validator"
	"dns-resolver/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository заменяет реальный репозиторий
type MockRepository struct {
	models.Repository 
}

func (m *MockRepository) AddOrUpdate(ctx context.Context, fqdn, ip string) error {
	return nil // Просто возвращаем успех
}

func (m *MockRepository) GetFQDNsByIP(ctx context.Context, ip string) ([]string, error) {
	if ip == "1.1.1.1" {
		return []string{"example.com"}, nil
	}
	return nil, nil
}

func (m *MockRepository) GetIPsByFQDN(ctx context.Context, fqdn string) ([]string, error) {
	if fqdn == "example.com" {
		return []string{"1.1.1.1"}, nil
	}
	return nil, nil
}

func TestAPIHandlers(t *testing.T) {
	//Создаем мок репозитория
	mockRepo := &MockRepository{}

	//Инициализируем реальный Resolver с моком репозитория
	resolver := dnsresolver.NewResolver(mockRepo)

	//Создаем обработчики API
	e := echo.New()
	e.Validator = v.New()

	h := NewHandler(resolver)
	h.RegisterRoutes(e)

	t.Run("AddFQDN success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/fqdns",
			strings.NewReader(`{"fqdn":"example.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := h.AddFQDN(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"fqdn":"example.com"`)
	})

	t.Run("AddFQDN - wrong method GET", func(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/api/fqdns", nil)
    rec := httptest.NewRecorder()
    
    e.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("GetFQDNsByIP success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=1.1.1.1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetFQDNsByIP(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"ip":"1.1.1.1","fqdns":["example.com"]}`, rec.Body.String())
	})
	

	t.Run("GetIPsByFQDN success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ips?fqdn=example.com", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetIPsByFQDN(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"fqdn":"example.com", "ips":["1.1.1.1"]}`, rec.Body.String())
	})
}

