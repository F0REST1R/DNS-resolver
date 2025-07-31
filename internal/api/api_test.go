package api

import (
	"context"
	"dns-resolver/internal/config"
	dnsresolver "dns-resolver/internal/dns_resolver"
	"dns-resolver/internal/models"
	"dns-resolver/internal/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockResolver подменяет реальный DNS-резолвер для тестов
type MockResolver struct {
	realResolver *dnsresolver.Resolver
}

func (m *MockResolver) Resolve(ctx context.Context, fqdn string) ([]string, error) {
	// Возвращаем фиксированные IP для тестовых доменов
	switch fqdn {
	case "github.com.":
		return []string{"140.82.121.4"}, nil
	case "support.microsoft.com":
		return []string{"104.215.148.63", "40.76.4.15"}, nil
	default:
		return m.realResolver.Resolve(ctx, fqdn)
	}
}

// Остальные методы проксируем к реальному резолверу
func (m *MockResolver) GetFQDNsByIP(ctx context.Context, ip string) ([]string, error) {
	return m.realResolver.GetFQDNsByIP(ctx, ip)
}

func (m *MockResolver) GetIPsByFQDN(ctx context.Context, fqdn string) ([]string, error) {
	return m.realResolver.GetIPsByFQDN(ctx, fqdn)
}

func (m *MockResolver) DNSUpdater(ctx context.Context, interval time.Duration) {
	m.realResolver.DNSUpdater(ctx, interval)
}

func setupTestDB(t *testing.T) *repository.DB {
	db, err := config.TestDBcon()
	require.NoError(t, err)

	// Очищаем и мигрируем тестовую БД
	err = db.Exec("DROP TABLE IF EXISTS dns_records").Error
	require.NoError(t, err)
	err = db.AutoMigrate(&models.DNSRecord{})
	require.NoError(t, err)

	return repository.NewDB(db)
}

func TestAPIWithRealDB(t *testing.T) {
	// Инициализация тестовой БД
	db := setupTestDB(t)
	resolver := dnsresolver.NewResolver(db)
	h := NewHandler(resolver)

	e := echo.New()
	e.Validator = &Validator{validator: NewValidator()}
	h.RegisterRoutes(e)

	ctx := context.Background()

	t.Run("POST /api/fqdns - успешное добавление", func(t *testing.T) {
		body := `{"fqdn":"github.com."}`
		req := httptest.NewRequest(http.MethodPost, "/api/fqdns", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"fqdn":"github.com."`)
		assert.Contains(t, rec.Body.String(), `"ips"`) // Проверяем что IP были получены
	})

	t.Run("GET /api/fqdns?ip=... - поиск по IP", func(t *testing.T) {
		// Подготовка данных
		err := db.AddOrUpdate(ctx, "github.com.", "140.82.121.4")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=140.82.121.4", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"fqdns":["github.com."]`)
	})

	t.Run("GET /api/ips?fqdn=github.com. - поиск по FQDN", func(t *testing.T) {
		err := db.AddOrUpdate(ctx, "github.com.", "140.82.121.4")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/ips?fqdn=github.com.", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
		assert.Equal(t, "github.com.", response["fqdn"])
		assert.NotEmpty(t, response["ips"])
		assert.IsType(t, []interface{}{}, response["ips"])
	})

	t.Run("POST /api/fqdns - неверный формат запроса", func(t *testing.T) {
		body := `{"invalid":"data"}`
		req := httptest.NewRequest(http.MethodPost, "/api/fqdns", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("POST /api/fqdns - FQDN with multiple IPs", func(t *testing.T) {
		body := `{"fqdn":"support.microsoft.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/fqdns", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		// Проверяем что вернулось несколько IP
		assert.Contains(t, rec.Body.String(), `"ips":["`)
		assert.Contains(t, rec.Body.String(), `","`)
	})

	t.Run("GET /api/fqdns?ip=... - SQL injection attempt", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/fqdns", nil)
		q := req.URL.Query()
		q.Add("ip", "' OR '1'='1")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"fqdns":[]`)
	})

	t.Run("GET /api/fqdns?ip=... - multiple FQDNs", func(t *testing.T) {
		// Подготовка данных
		err := db.AddOrUpdate(ctx, "www.cloudflare.com", "104.16.85.20")
		require.NoError(t, err)
		err = db.AddOrUpdate(ctx, "www.digitalocean.com", "104.16.85.20")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=104.16.85.20", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `www.cloudflare.com`)
		assert.Contains(t, rec.Body.String(), `www.digitalocean.com`)
	})

	t.Run("GET /api/fqdns?ip=... - IP не найден", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=127.0.0.1", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"ip":"127.0.0.1","fqdns":[]}`, rec.Body.String())
	})
}
