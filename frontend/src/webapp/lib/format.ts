// Форматтеры показаний для WebApp. Локаль ru-RU; деньги/вес выводим как есть (бэкенд уже
// дал готовое decimal-значение строкой), здесь только оформление.

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

// formatMoney форматирует decimal-строку напрямую, НЕ через Number: бэкенд отдаёт деньги
// строкой именно чтобы не терять точность, а Number('1000000.50') округлил бы до 1000000.5
// (потеря копеек) и поплыл бы за пределами ~15 значащих цифр.
export function formatMoney(amount: string | null, currency: string): string | null {
  if (!amount) {
    return null
  }

  const sign = amount.startsWith('-') ? '-' : ''
  const unsigned = sign ? amount.slice(1) : amount
  const [intPart, fracPart] = unsigned.split('.')

  // Не число — отдаём как есть с валютой (страховка от неожиданного формата).
  if (!/^\d+$/.test(intPart) || (fracPart !== undefined && !/^\d+$/.test(fracPart))) {
    return `${amount} ${currency}`
  }

  // Разряды через неразрывный пробел (как ru-RU группирует тысячи), дробную часть — как есть.
  const grouped = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
  const value = fracPart !== undefined ? `${grouped},${fracPart}` : grouped

  return `${sign}${value} ${currency}`
}

export function formatWeight(weight: string | null): string | null {
  return weight ? `${weight} кг` : null
}

export function formatVolume(volume: string | null): string | null {
  return volume ? `${volume} м³` : null
}
