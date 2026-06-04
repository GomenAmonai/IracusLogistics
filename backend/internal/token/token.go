package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
)

// Claims — полезная нагрузка JWT. Помимо стандартных claims несёт Role, чтобы один и тот
// же секрет подписи различал токены менеджера и клиента. Без Role клиентский токен
// (валидная подпись, Subject = id клиента) прошёл бы менеджерскую проверку — и клиент
// получил бы доступ к менеджерским ручкам. Subject = id субъекта (менеджера или клиента).
type Claims struct {
	jwt.RegisteredClaims
	Role domain.Role `json:"role"`
}

// Issue подписывает HS256-токен с субъектом, ролью и сроком жизни ttl. Единственное место
// выпуска токенов — и менеджерских, и клиентских, чтобы форма claims не разъезжалась.
func Issue(secret []byte, subject uuid.UUID, role domain.Role, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Role: role,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// Parse проверяет подпись и срок токена и возвращает claims. Принимает только HS256
// (защита от alg-подмены: иначе можно подсунуть "none"/RS256) и требует exp (иначе
// корректно подписанный токен без exp жил бы вечно). Проверку роли делает вызывающий.
func Parse(secret []byte, raw string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return secret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}
