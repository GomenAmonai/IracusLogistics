package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Параметры самоочистки таблицы посетителей: запись старше visitorTTL удаляется при
// ближайшем проходе sweep (не чаще раза в sweepInterval). Так таблица не растёт безгранично
// под перебором IP, и при этом не нужна фоновая горутина с отдельным lifecycle.
const (
	visitorTTL    = 10 * time.Minute
	sweepInterval = time.Minute
)

// RateLimit ограничивает частоту запросов по IP (token bucket: rps токенов/сек, ёмкость burst).
// Превышение → 429 в общем формате ошибок. Вешается на публичные ручки (лиды, логин,
// Telegram-авторизация) как защита от спама и перебора.
//
// NOTE: MVP — IP берём из gin c.ClientIP() (доверяет X-Forwarded-For); за неизвестным прокси
// заголовок можно подделать. Ужесточить trusted-proxies при деплое; см. docs/tech-debt.md
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := &rateLimiter{
		visitors:  make(map[string]*visitor),
		rate:      rps,
		burst:     float64(burst),
		lastSweep: time.Now(),
	}

	return func(c *gin.Context) {
		if !limiter.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "rate_limited", "message": "Слишком много запросов, попробуйте позже"},
			})
			return
		}

		c.Next()
	}
}

type visitor struct {
	tokens float64
	last   time.Time
}

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*visitor
	rate      float64
	burst     float64
	lastSweep time.Time
}

func (l *rateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.sweep(now)

	v, ok := l.visitors[ip]
	if !ok {
		l.visitors[ip] = &visitor{tokens: l.burst - 1, last: now}
		return true
	}

	// Пополняем корзину пропорционально прошедшему времени, но не выше ёмкости.
	v.tokens = min(l.burst, v.tokens+now.Sub(v.last).Seconds()*l.rate)
	v.last = now
	if v.tokens < 1 {
		return false
	}
	v.tokens--

	return true
}

func (l *rateLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < sweepInterval {
		return
	}
	l.lastSweep = now
	for ip, v := range l.visitors {
		if now.Sub(v.last) > visitorTTL {
			delete(l.visitors, ip)
		}
	}
}
