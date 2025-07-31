package repository

import (
	"context"
	"dns-resolver/internal/config"
	"dns-resolver/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	db, err := config.TestDBcon()
	require.NoError(t, err)

	err = db.Exec("DROP TABLE IF EXISTS dns_records").Error
	require.NoError(t, err)

	err = db.AutoMigrate(&models.DNSRecord{})
	require.NoError(t, err, "Failed to migrate test database")

	repo := NewDB(db)
	ctx := context.Background()

	t.Run("GetIPsByFQDN", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)

		repo.AddOrUpdate(ctx, "example.com", "1.1.1.1")
		repo.AddOrUpdate(ctx, "example.com", "1.1.2.2")

		ips, err := repo.GetIPsByFQDN(ctx, "example.com")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"1.1.1.1", "1.1.2.2"}, ips)
	})

	t.Run("GetIPsByFQDN empty result", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)

		ips, err := repo.GetIPsByFQDN(ctx, "nonexistent.com")
		require.NoError(t, err)
		assert.Empty(t, ips)
	})

	t.Run("GetFQDNsByIP", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)
		// Подготовка данных
		repo.AddOrUpdate(ctx, "site1.com", "3.3.3.3")
		repo.AddOrUpdate(ctx, "site2.com", "3.3.3.3")

		// Тестирование
		fqdns, err := repo.GetFQDNsByIP(ctx, "3.3.3.3")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"site1.com", "site2.com"}, fqdns)
	})

	t.Run("GetFQDNsByIP empty result", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)

		fqdns, err := repo.GetFQDNsByIP(ctx, "0.0.0.0")
		require.NoError(t, err)
		assert.Empty(t, fqdns)
	})

	t.Run("AddOrUpdate creates new record", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)

		err = repo.AddOrUpdate(ctx, "new.com", "5.5.5.5")
		require.NoError(t, err)

		// Проверяем что запись действительно создалась
		var count int64
		db.Model(&models.DNSRecord{}).Where("fqdn = ? AND ip = ?", "new.com", "5.5.5.5").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("AddOrUpdate updates existing record", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)
		// Сначала создаем запись
		require.NoError(t, repo.AddOrUpdate(ctx, "exist.com", "6.6.6.6"))

		// Получаем оригинальное время создания
		var original models.DNSRecord
		db.Where("fqdn = ? AND ip = ?", "exist.com", "6.6.6.6").First(&original)

		// Обновляем запись
		require.NoError(t, repo.AddOrUpdate(ctx, "exist.com", "6.5.5.6"))

		// Проверяем что updated_at изменился
		var updated models.DNSRecord
		db.Where("fqdn = ? AND ip = ?", "exist.com", "6.5.5.6").First(&updated)
		assert.True(t, updated.UpdatedAt.After(original.UpdatedAt))
	})

	t.Run("GetAllFQDNs", func(t *testing.T) {
		err := db.Exec("DELETE FROM dns_records").Error
		require.NoError(t, err)
		repo.AddOrUpdate(ctx, "site1.com", "3.3.3.3")
		repo.AddOrUpdate(ctx, "site2.com", "3.2.2.3")

		fqdns, err := repo.GetAllFQDNs(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"site1.com", "site2.com"}, fqdns)
	})
}
