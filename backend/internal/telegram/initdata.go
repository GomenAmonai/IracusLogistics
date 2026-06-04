// Package telegram содержит проверку подлинности данных Telegram WebApp (initData).
package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidInitData — initData не прошёл проверку (битая подпись, подделка, протух).
var ErrInvalidInitData = errors.New("invalid telegram init data")

// User — пользователь Telegram из поля user в initData. Поля 1:1 с JSON Telegram.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// DisplayName собирает отображаемое имя: «Имя Фамилия», иначе username, иначе «Клиент».
// Client.name — NOT NULL, поэтому пустым он быть не должен.
func (u User) DisplayName() string {
	name := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if name != "" {
		return name
	}
	if u.Username != "" {
		return u.Username
	}

	return "Клиент"
}

// ValidateInitData проверяет подпись initData по алгоритму Telegram WebApp и свежесть
// auth_date, затем возвращает пользователя. botToken — секрет бота; пустой токен → ошибка
// (без секрета подделку не отличить, поэтому отказываем). maxAge — окно свежести (0 — не
// проверять). Алгоритм: secret_key = HMAC_SHA256(key="WebAppData", msg=botToken);
// hash == HMAC_SHA256(key=secret_key, msg=data_check_string), где data_check_string —
// все поля кроме hash, отсортированные по ключу и склеенные "key=value" через \n.
func ValidateInitData(initData, botToken string, maxAge time.Duration) (*User, error) {
	if botToken == "" {
		return nil, fmt.Errorf("%w: bot token not configured", ErrInvalidInitData)
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("%w: parse query: %v", ErrInvalidInitData, err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("%w: missing hash", ErrInvalidInitData)
	}
	values.Del("hash")

	if !validSignature(values, botToken, hash) {
		return nil, fmt.Errorf("%w: signature mismatch", ErrInvalidInitData)
	}

	if maxAge > 0 {
		authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: bad auth_date", ErrInvalidInitData)
		}
		if time.Since(time.Unix(authDate, 0)) > maxAge {
			return nil, fmt.Errorf("%w: expired", ErrInvalidInitData)
		}
	}

	rawUser := values.Get("user")
	if rawUser == "" {
		return nil, fmt.Errorf("%w: missing user", ErrInvalidInitData)
	}
	var user User
	if err := json.Unmarshal([]byte(rawUser), &user); err != nil {
		return nil, fmt.Errorf("%w: parse user: %v", ErrInvalidInitData, err)
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("%w: missing user id", ErrInvalidInitData)
	}

	return &user, nil
}

// validSignature собирает data_check_string и сравнивает HMAC за постоянное время.
func validSignature(values url.Values, botToken, hash string) bool {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(values.Get(k))
	}

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))

	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(sb.String()))
	computed := hex.EncodeToString(mac.Sum(nil))

	// hmac.Equal — сравнение за постоянное время (защита от тайминг-атак на hash).
	return hmac.Equal([]byte(computed), []byte(hash))
}
