package domain

import "errors"

// ErrNotFound — общий «не найдено». Объявлен в domain, потому что это слово из общего
// словаря: его возвращает repository и понимает service, а оба и так зависят от domain.
// Так service не вынужден импортировать repository ради одной ошибки.
var ErrNotFound = errors.New("not found")
