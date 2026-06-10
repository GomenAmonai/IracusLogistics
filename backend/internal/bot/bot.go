package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
)

// Bot отправляет уведомления и принимает команды клиентов в Telegram. Структурно
// удовлетворяет интерфейсам-уведомителям сервисов (Notifier, ClientNotifier, ...), но
// пакет service не импортирует — связь только по контракту методов.
//
// NOTE: MVP — уведомление о лидах идёт в один чат через chatID (не пер-менеджер); бот на
// long polling; см. docs/tech-debt.md
type Bot struct {
	api    *tgbotapi.BotAPI
	chatID int64
	// disabled — режим no-op: токен не задан, ничего не шлём и не принимаем (dev без токена).
	disabled bool
	// outbox — персистентная очередь уведомлений (см. UseOutbox); nil => прямая отправка.
	outbox OutboxStore
	// webhookDeps — зависимости обработки апдейтов в webhook-режиме (см. StartWebhook);
	// nil => апдейты приходят через long polling (Run).
	webhookDeps *RunDeps
}

// ClientRegistrar создаёт/находит клиента по его Telegram-личности (реализует
// service.ClientService). Бот зовёт его из обработчика /start.
type ClientRegistrar interface {
	Register(ctx context.Context, telegramID int64, username, name string, leadID *uuid.UUID) (*domain.Client, error)
}

// ShipmentLister отдаёт грузы клиента по telegram_id (реализует service.ShipmentService).
// Бот зовёт его из обработчика /status.
type ShipmentLister interface {
	ListByTelegramID(ctx context.Context, telegramID int64) ([]domain.Shipment, error)
}

// RunDeps — то, что нужно циклу обработки команд. Передаётся в Run, а не в New, потому что
// сервисы конструируются после бота (бот сам — их уведомитель).
type RunDeps struct {
	Registrar ClientRegistrar
	Lister    ShipmentLister
}

// New создаёт бот. Пустой token => рабочий no-op Bot (для dev без токена): он логирует
// "bot disabled" один раз и молча проглатывает уведомления и команды. Пустой chatID при
// заданном token — ошибка конфигурации.
//
// NOTE: при заданном token New синхронно вызывает getMe (валидация токена), то есть
// требует доступности Telegram на старте.
func New(token, chatID string) (*Bot, error) {
	if token == "" {
		log.Print("bot disabled: TELEGRAM_BOT_TOKEN not set, notifications and commands are no-op")
		return &Bot{disabled: true}, nil
	}

	if chatID == "" {
		return nil, fmt.Errorf("bot: chat id is required when token is set")
	}

	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bot: parse chat id: %w", err)
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		// Ошибка библиотеки может содержать URL с токеном, поэтому наружу её не отдаём.
		// Но категорию сбоя (тип ошибки) залогировать безопасно — иначе диагностики ноль.
		slog.Error("bot: init telegram api", "kind", fmt.Sprintf("%T", err))
		return nil, fmt.Errorf("bot: init telegram api failed")
	}

	return &Bot{api: api, chatID: id}, nil
}

// Run запускает long polling и обрабатывает команды клиентов до отмены ctx. Блокирующий —
// вызывается в отдельной горутине. В no-op режиме сразу возвращается. Альтернативный
// источник апдейтов — webhook (StartWebhook + ProcessUpdateJSON), тогда Run не вызывается.
func (b *Bot) Run(ctx context.Context, deps RunDeps) {
	if b.disabled {
		return
	}

	// Зарегистрированный ранее webhook блокирует getUpdates (Telegram отвечает 409),
	// поэтому перед поллингом снимаем его. Ошибку только логируем: если webhook не был
	// настроен, удалять нечего, а поллинг всё равно стартует.
	if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		slog.Warn("bot: delete webhook before polling", "kind", fmt.Sprintf("%T", err))
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update := <-updates:
			b.processUpdate(ctx, deps, update)
		}
	}
}

// processUpdate — обработка одного апдейта; общая для поллинга и webhook.
func (b *Bot) processUpdate(ctx context.Context, deps RunDeps, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, deps, update.CallbackQuery)
		return
	}
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}
	b.handleCommand(ctx, deps, update.Message)
}

func (b *Bot) handleCommand(ctx context.Context, deps RunDeps, msg *tgbotapi.Message) {
	// Паника при обработке одного апдейта не должна валить цикл и весь процесс: бот крутится
	// в bare-горутине вне gin.Recovery, поэтому ловим её здесь.
	defer func() {
		if r := recover(); r != nil {
			slog.Error("bot: panic handling command", "recover", r)
		}
	}()

	// Сообщение без отправителя (анонимный админ группы, автопересылка) не несёт личности
	// для регистрации/статуса — обе команды требуют msg.From.ID, поэтому игнорируем; иначе
	// разыменование nil уронило бы процесс.
	if msg.From == nil {
		return
	}

	switch msg.Command() {
	case "start":
		b.handleStart(ctx, deps, msg)
	case "menu":
		b.replyMenu(msg.Chat.ID, "Чем помочь?")
	case "status":
		b.handleStatus(ctx, deps, msg)
	default:
		b.replyMenu(msg.Chat.ID, "Не знаю такую команду. Вот меню:")
	}
}

// handleCallback обрабатывает нажатия inline-кнопок. Сначала «гасит» крутилку на кнопке
// (AnswerCallbackQuery), затем выполняет действие. Защищён recover'ом и nil-проверкой
// отправителя — как handleCommand (паника в bare-горутине уронила бы процесс).
func (b *Bot) handleCallback(ctx context.Context, deps RunDeps, cb *tgbotapi.CallbackQuery) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("bot: panic handling callback", "recover", r)
		}
	}()

	b.answerCallback(cb.ID)

	if cb.From == nil || cb.Message == nil {
		return
	}
	chatID := cb.Message.Chat.ID

	switch cb.Data {
	case "status":
		b.sendStatus(ctx, deps, chatID, cb.From.ID)
	case "help":
		b.replyMenu(chatID, helpText())
	case "manager":
		b.reply(chatID, managerContactText())
	default:
		b.replyMenu(chatID, "Неизвестная кнопка. Вот меню:")
	}
}

// handleStart регистрирует клиента по его Telegram-личности. Аргумент команды (deep-link
// payload «?start=<lead_id>») при наличии привязывает клиента к исходной заявке.
func (b *Bot) handleStart(ctx context.Context, deps RunDeps, msg *tgbotapi.Message) {
	leadID := parseLeadID(msg.CommandArguments())

	_, err := deps.Registrar.Register(ctx, msg.From.ID, msg.From.UserName, displayName(msg.From), leadID)
	if err != nil {
		slog.Error("bot: register client", "telegram_id", msg.From.ID, "error", err)
		b.reply(msg.Chat.ID, "Не удалось завершить регистрацию. Попробуйте позже.")
		return
	}

	b.replyMenu(msg.Chat.ID, "Аккаунт подтверждён. Чем помочь?")
}

func (b *Bot) handleStatus(ctx context.Context, deps RunDeps, msg *tgbotapi.Message) {
	b.sendStatus(ctx, deps, msg.Chat.ID, msg.From.ID)
}

// sendStatus отправляет список грузов клиента по telegram_id. Общий путь для команды
// /status и кнопки «Мои грузы».
func (b *Bot) sendStatus(ctx context.Context, deps RunDeps, chatID, telegramID int64) {
	shipments, err := deps.Lister.ListByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			b.reply(chatID, "Вы ещё не зарегистрированы. Отправьте /start.")
			return
		}
		slog.Error("bot: list shipments", "telegram_id", telegramID, "error", err)
		b.reply(chatID, "Не удалось получить грузы. Попробуйте позже.")
		return
	}

	b.reply(chatID, formatStatusList(shipments))
}

// NotifyNewLead шлёт уведомление о новом лиде в чат менеджера.
//
// NOTE: MVP — библиотека v5 не принимает context, поэтому ctx не ограничивает таймаут и не
// отменяет прямую отправку; в режиме outbox ctx работает на постановке в очередь.
func (b *Bot) NotifyNewLead(ctx context.Context, lead *domain.Lead) error {
	return b.deliver(ctx, "lead_new", b.chatID, formatLeadMessage(lead))
}

// NotifyShipmentStatus уведомляет клиента о смене статуса груза (шлём на его telegram_id).
func (b *Bot) NotifyShipmentStatus(ctx context.Context, telegramID int64, shipment *domain.Shipment) error {
	return b.deliver(ctx, "shipment_status", telegramID, formatStatusUpdate(shipment))
}

// NotifyClientMessage уведомляет менеджера о новом сообщении клиента.
func (b *Bot) NotifyClientMessage(ctx context.Context, client *domain.Client, shipment *domain.Shipment, text string) error {
	return b.deliver(ctx, "client_message", b.chatID, formatClientMessage(client, shipment, text))
}

// NotifyManagerReply уведомляет клиента об ответе менеджера (шлём на его telegram_id).
func (b *Bot) NotifyManagerReply(ctx context.Context, telegramID int64, shipment *domain.Shipment, text string) error {
	return b.deliver(ctx, "manager_reply", telegramID, formatManagerReply(shipment, text))
}

// send — единая точка отправки: no-op в disabled-режиме, ошибку отдаёт без первопричины
// (строка из Send содержит URL с токеном в пути — её нельзя пробрасывать в лог).
func (b *Bot) send(chatID int64, text string) error {
	if b.disabled {
		return nil
	}

	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("bot: send message failed")
	}

	return nil
}

// reply отправляет ответ на команду; ошибку только логирует (не роняем цикл обработки).
func (b *Bot) reply(chatID int64, text string) {
	if err := b.send(chatID, text); err != nil {
		slog.Error("bot: reply failed", "chat_id", chatID, "error", err)
	}
}

// landingURL — публичный сайт, на который ведёт кнопка меню. Mini App открывается отдельной
// кнопкой-меню Telegram (setChatMenuButton), web_app-кнопки в inline нет в v5.5.1.
const landingURL = "https://icaris-logistics.vercel.app"

// mainMenu — базовое inline-меню по функционалу клиента.
func mainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Мои грузы", "status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Как это работает", "help"),
			tgbotapi.NewInlineKeyboardButtonData("👤 Менеджер", "manager"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🌐 Сайт", landingURL),
		),
	)
}

// replyMenu отправляет текст с основным inline-меню; ошибку только логирует.
func (b *Bot) replyMenu(chatID int64, text string) {
	if b.disabled {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainMenu()
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("bot: reply menu failed", "chat_id", chatID)
	}
}

// answerCallback «гасит» крутилку на нажатой кнопке (Telegram требует ответить на callback).
func (b *Bot) answerCallback(id string) {
	if b.disabled {
		return
	}
	if _, err := b.api.Request(tgbotapi.NewCallback(id, "")); err != nil {
		slog.Error("bot: answer callback failed")
	}
}

// parseLeadID разбирает аргумент /start как UUID лида. Пустой/битый аргумент → nil (клиент
// пришёл без deep-link).
func parseLeadID(arg string) *uuid.UUID {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return nil
	}
	id, err := uuid.Parse(arg)
	if err != nil {
		return nil
	}

	return &id
}

func displayName(user *tgbotapi.User) string {
	if user == nil {
		return ""
	}
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name != "" {
		return name
	}

	return user.UserName
}
