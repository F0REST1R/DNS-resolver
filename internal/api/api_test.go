package api

import (
	"context"
	"dns-resolver/internal/config"
	dnsresolver "dns-resolver/internal/dns_resolver"
	"dns-resolver/internal/models"
	"dns-resolver/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		body := `{"fqdn":"example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/fqdns", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"fqdn":"example.com"`)
		assert.Contains(t, rec.Body.String(), `"ips"`) // Проверяем что IP были получены
	})

	t.Run("GET /api/fqdns?ip=... - поиск по IP", func(t *testing.T) {
		// Подготовка данных
		err := db.AddOrUpdate(ctx, "test.com", "8.8.8.8")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=8.8.8.8", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"fqdns":["test.com"]`)
	})

	t.Run("GET /api/IP/:fqdn - поиск по FQDN", func(t *testing.T) {
		err := db.AddOrUpdate(ctx, "test2.com", "1.1.1.1")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/ips?fqdn=test2.com", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"ips":["1.1.1.1"]`)
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
		body := `{"fqdn":"multi-ip.com"}`
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
		err := db.AddOrUpdate(ctx, "site1.com", "9.9.9.9")
		require.NoError(t, err)
		err = db.AddOrUpdate(ctx, "site2.com", "9.9.9.9")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=9.9.9.9", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `site1.com`)
		assert.Contains(t, rec.Body.String(), `site2.com`)
	})

	t.Run("GET /api/fqdns?ip=... - IP не найден", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/fqdns?ip=127.0.0.1", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"ip":"127.0.0.1","fqdns":[]}`, rec.Body.String())
	})
}
