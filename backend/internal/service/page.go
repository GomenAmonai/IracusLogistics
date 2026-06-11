package service

// Page — параметры пагинации списков. Нормализуется сервисом (не хендлером), чтобы
// границы действовали для любого вызывающего, включая будущие CLI/бот-пути.
type Page struct {
	Limit  int
	Offset int
}

const (
	defaultPageLimit = 100
	maxPageLimit     = 200
)

// normalize приводит границы: невалидный/нулевой limit → дефолт, потолок — maxPageLimit,
// отрицательный offset → 0. Потолок защищает БД от полного скана (техдолг #25).
func (p Page) normalize() Page {
	if p.Limit <= 0 {
		p.Limit = defaultPageLimit
	}
	if p.Limit > maxPageLimit {
		p.Limit = maxPageLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	return p
}
