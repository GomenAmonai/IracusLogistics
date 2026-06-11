package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/telegram"
	"icaris-logistic/backend/internal/token"
)

// ErrTelegramAuth — initData не прошёл проверку подлинности (подделка/протух). Хендлер
// отдаёт по нему 401, отдельно от ErrValidation (400), чтобы фронт мог различить.
var ErrTelegramAuth = errors.New("telegram authentication failed")

// initDataMaxAge — окно свежести initData: подпись Telegram старше окна отклоняем (защита
// от воспроизведения перехваченной строки). NOTE: MVP — фиксированное окно; см. docs/tech-debt.md
const initDataMaxAge = 24 * time.Hour

// ClientStore — то, что ClientService нужно от хранилища клиентов (интерфейс на стороне
// потребителя, как ManagerStore/LeadStore).
type ClientStore interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error)
	Create(ctx context.Context, client *domain.Client) error
	List(ctx context.Context, limit, offset int) ([]domain.Client, error)
}

type ClientService struct {
	clients  ClientStore
	botToken string
	secret   []byte
	ttl      time.Duration
}

func NewClientService(clients ClientStore, botToken, secret string, ttl time.Duration) *ClientService {
	return &ClientService{clients: clients, botToken: botToken, secret: []byte(secret), ttl: ttl}
}

// Register находит клиента по telegram_id или создаёт нового — идемпотентно. Общий путь
// для WebApp-авторизации и команды /start в боте. leadID привязывает клиента к исходной
// заявке (учитывается только при первом создании; повторный /start существующего клиента
// его не перепривязывает).
func (s *ClientService) Register(
	ctx context.Context,
	telegramID int64,
	username, name string,
	leadID *uuid.UUID,
) (*domain.Client, error) {
	if telegramID == 0 {
		return nil, fmt.Errorf("%w: telegram id is required", ErrValidation)
	}

	existing, err := s.clients.GetByTelegramID(ctx, telegramID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "Клиент" // name NOT NULL; страхуемся, хотя у Telegram first_name почти всегда есть
	}

	client := &domain.Client{
		TelegramID: telegramID,
		Username:   strings.TrimSpace(username),
		Name:       name,
		LeadID:     leadID,
	}
	if err := s.clients.Create(ctx, client); err != nil {
		switch {
		case errors.Is(err, domain.ErrClientExists):
			// Гонка по telegram_id: параллельный Register тем же telegram_id уже создал клиента.
			// Перечитываем — это нормальный исход, а не ошибка.
			if recovered, getErr := s.clients.GetByTelegramID(ctx, telegramID); getErr == nil {
				return recovered, nil
			}
			return nil, fmt.Errorf("register: re-read after telegram_id conflict: %w", err)

		case errors.Is(err, domain.ErrLeadAlreadyClaimed) && client.LeadID != nil:
			// Этот lead уже привязан к другому клиенту (двое открыли один deep-link
			// /start=<lead_id>). Не блокируем регистрацию навсегда — заводим клиента без привязки
			// к заявке (leadID привязывается best-effort).
			client.LeadID = nil
			if retryErr := s.clients.Create(ctx, client); retryErr != nil {
				return nil, fmt.Errorf("register: create after unbinding lead: %w", retryErr)
			}
			return client, nil

		default:
			// Любая прочая ошибка (обрыв соединения, дедлок, NOT NULL) — НЕ маскируем под гонку и
			// НЕ создаём клиента без привязки к заявке. Пробрасываем.
			return nil, fmt.Errorf("register: create client: %w", err)
		}
	}

	return client, nil
}

// AuthenticateWebApp проверяет подпись initData Telegram, регистрирует/находит клиента и
// выдаёт client-JWT для запросов WebApp.
func (s *ClientService) AuthenticateWebApp(ctx context.Context, initData string) (string, *domain.Client, error) {
	user, err := telegram.ValidateInitData(initData, s.botToken, initDataMaxAge)
	if err != nil {
		// Не пробрасываем первопричину наружу как 500 — это ожидаемый отказ авторизации.
		return "", nil, fmt.Errorf("%w: %v", ErrTelegramAuth, err)
	}

	client, err := s.Register(ctx, user.ID, user.Username, user.DisplayName(), nil)
	if err != nil {
		return "", nil, err
	}

	signed, err := token.Issue(s.secret, client.ID, domain.RoleClient, s.ttl)
	if err != nil {
		return "", nil, err
	}

	return signed, client, nil
}

func (s *ClientService) List(ctx context.Context, page Page) ([]domain.Client, error) {
	page = page.normalize()
	return s.clients.List(ctx, page.Limit, page.Offset)
}
