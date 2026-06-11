package bot

import (
	"strings"

	"icaris-logistic/backend/internal/domain"
)

// statusLabels — русские подписи статусов груза для сообщений бота. Отдельная карта в слое
// бота: presentation-строки не место в domain. Фронтенд держит свой набор подписей.
var statusLabels = map[domain.ShipmentStatus]string{
	domain.ShipmentStatusPending:        "Создан",
	domain.ShipmentStatusPickedUp:       "Забран у отправителя",
	domain.ShipmentStatusInTransit:      "В пути",
	domain.ShipmentStatusCustomsClear:   "Таможенное оформление",
	domain.ShipmentStatusInWarehouse:    "На складе",
	domain.ShipmentStatusOutForDelivery: "Доставляется",
	domain.ShipmentStatusDelivered:      "Доставлен",
	domain.ShipmentStatusCancelled:      "Отменён",
}

func statusLabel(status domain.ShipmentStatus) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}

	return string(status)
}

// formatLeadMessage собирает русское сообщение о лиде. Чистая функция без сети —
// тестируется напрямую. Вес/объём показываем только когда NullDecimal валиден.
func formatLeadMessage(lead *domain.Lead) string {
	var b strings.Builder

	b.WriteString("Новый лид с сайта\n\n")
	b.WriteString("Имя: " + lead.Name + "\n")
	b.WriteString("Телефон: " + lead.Phone + "\n")
	b.WriteString("Маршрут: " + lead.FromCity + " → " + lead.ToCity + "\n")

	if lead.Weight.Valid {
		b.WriteString("Вес: " + lead.Weight.Decimal.String() + " кг\n")
	}
	if lead.Volume.Valid {
		b.WriteString("Объём: " + lead.Volume.Decimal.String() + " м³\n")
	}
	if lead.CargoType != "" {
		b.WriteString("Тип груза: " + lead.CargoType + "\n")
	}
	if lead.Comment != "" {
		b.WriteString("Комментарий: " + lead.Comment + "\n")
	}

	b.WriteString("ID: " + lead.ID.String())

	return b.String()
}

// formatStatusList — ответ на /status: список грузов клиента с трек-ключом и статусом.
func formatStatusList(shipments []domain.Shipment) string {
	if len(shipments) == 0 {
		return "У вас пока нет грузов в работе."
	}

	var b strings.Builder
	b.WriteString("Ваши грузы:\n")
	for _, s := range shipments {
		b.WriteString("\n" + s.TrackingKey + " — " + statusLabel(s.Status))
		if s.FromCity != "" || s.ToCity != "" {
			b.WriteString("\n" + s.FromCity + " → " + s.ToCity)
		}
		if s.StatusComment != "" {
			b.WriteString("\n" + s.StatusComment)
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// formatStatusUpdate — пуш клиенту при смене статуса конкретного груза.
func formatStatusUpdate(shipment *domain.Shipment) string {
	var b strings.Builder
	b.WriteString("Статус груза " + shipment.TrackingKey + " обновлён: " + statusLabel(shipment.Status))
	if shipment.StatusComment != "" {
		b.WriteString("\n" + shipment.StatusComment)
	}

	return b.String()
}

// channelLabels — русские подписи каналов оплаты (как statusLabels — presentation-строки
// живут в слое бота).
var channelLabels = map[domain.PaymentChannel]string{
	domain.PaymentChannelBankTransfer: "безнал по счёту",
	domain.PaymentChannelCardSBP:      "карта / СБП",
	domain.PaymentChannelCash:         "наличные",
	domain.PaymentChannelCrypto:       "криптовалюта",
}

func channelLabel(channel domain.PaymentChannel) string {
	if label, ok := channelLabels[channel]; ok {
		return label
	}

	return string(channel)
}

// formatPaymentCreated — пуш клиенту о выставленном счёте (платёж создан в pending).
func formatPaymentCreated(shipment *domain.Shipment, payment *domain.Payment) string {
	var b strings.Builder
	b.WriteString("По грузу " + shipment.TrackingKey + " выставлен счёт: " +
		payment.Amount.String() + " " + payment.Currency + " (" + channelLabel(payment.Channel) + ").")
	if payment.Comment != "" {
		b.WriteString("\n" + payment.Comment)
	}
	b.WriteString("\nДетали — в приложении, в карточке груза.")

	return b.String()
}

// formatPaymentConfirmed — пуш клиенту о полученном платеже.
func formatPaymentConfirmed(shipment *domain.Shipment, payment *domain.Payment) string {
	return "Платёж по грузу " + shipment.TrackingKey + " получен: " +
		payment.Amount.String() + " " + payment.Currency + " (" + channelLabel(payment.Channel) + "). Спасибо!"
}

// formatClientMessage — уведомление менеджеру о новом сообщении клиента.
func formatClientMessage(client *domain.Client, shipment *domain.Shipment, text string) string {
	return "Сообщение от клиента " + client.Name + " по грузу " + shipment.TrackingKey + ":\n" + text
}

// formatManagerReply — уведомление клиенту об ответе менеджера.
func formatManagerReply(shipment *domain.Shipment, text string) string {
	return "Менеджер по грузу " + shipment.TrackingKey + ":\n" + text
}

// helpText — справка для кнопки «Как это работает».
func helpText() string {
	return "Icaris Logistics — доставка грузов Китай → Россия.\n\n" +
		"Здесь вы:\n" +
		"• 📦 смотрите статусы своих грузов («Мои грузы» или /status);\n" +
		"• получаете уведомления при каждой смене статуса автоматически;\n" +
		"• открываете приложение (кнопка слева от поля ввода) — там детали груза, история и чат с менеджером.\n\n" +
		"Команды: /start · /status · /menu"
}

// managerContactText — контакты для кнопки «Менеджер».
func managerContactText() string {
	return "Связаться с менеджером:\n" +
		"• Telegram: @hikill8 — https://t.me/hikill8\n\n" +
		"Вопрос по конкретному грузу удобнее задать в приложении — в чате по этому грузу."
}
