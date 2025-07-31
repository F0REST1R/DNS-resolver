package dnsresolver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository реализует интерфейс Repository для тестов
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) AddOrUpdate(ctx context.Context, fqdn, ip string) error {
	args := m.Called(ctx, fqdn, ip)
	return args.Error(0)
}

func (m *MockRepository) GetIPsByFQDN(ctx context.Context, fqdn string) ([]string, error) {
	args := m.Called(ctx, fqdn)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) GetFQDNsByIP(ctx context.Context, ip string) ([]string, error) {
	args := m.Called(ctx, ip)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) GetAllFQDNs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func TestDNSUpdater(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := new(MockRepository)
	resolver := NewResolver(mockRepo)

	// Устанавливаем ожидания для мока
	testFqdns := []string{"example.com", "test.com"}
	mockRepo.On("GetAllFQDNs", mock.Anything).Return(testFqdns, nil)
	mockRepo.On("AddOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Запускаем обновление с маленьким интервалом
	go resolver.DNSUpdater(ctx, 100*time.Millisecond)

	// Ждем завершения контекста
	<-ctx.Done()

	// Проверяем что методы вызывались с правильными параметрами
	mockRepo.AssertCalled(t, "GetAllFQDNs", mock.Anything)
	for _, fqdn := range testFqdns {
		mockRepo.AssertCalled(t, "AddOrUpdate", mock.Anything, fqdn, mock.Anything)
	}
}

func TestDNSUpdater_ErrorHandling(t *testing.T) {
	// Создаем мок репозитория с ошибкой
	mockRepo := new(MockRepository)
	resolver := NewResolver(mockRepo)

	// Устанавливаем ошибку при получении FQDNs
	mockRepo.On("GetAllFQDNs", mock.Anything).Return([]string{}, assert.AnError)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Запускаем обновление
	go resolver.DNSUpdater(ctx, 50*time.Millisecond)

	// Ждем завершения
	<-ctx.Done()

	// Проверяем что AddOrUpdate не вызывался при ошибке
	mockRepo.AssertNotCalled(t, "AddOrUpdate", mock.Anything, mock.Anything, mock.Anything)
}