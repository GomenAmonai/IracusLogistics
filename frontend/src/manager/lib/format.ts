// Форматтеры панели — те же, что в клиентском WebApp (особенно formatMoney: деньги
// форматируются строкой, НЕ через Number — там обоснование). Единственная точка связи
// между приложениями; при расхождении нужд — скопировать и развести.
export { formatDate, formatDateTime, formatMoney, formatVolume, formatWeight } from '../../webapp/lib/format'
