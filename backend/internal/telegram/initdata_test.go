package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testBotToken = "123456:test-bot-token"

// signValues собирает initData с корректной подписью по алгоритму Telegram — тот же, что
// проверяет ValidateInitData, но независимая реализация в тесте: если в продакшен-коде
// собьётся сортировка/кодирование, round-trip перестанет сходиться.
func signValues(botToken string, fields map[string]string) url.Values {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+"="+fields[k])
	}

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(lines, "\n")))

	values := url.Values{}
	for k, v := range fields {
		values.Set(k, v)
	}
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))

	return values
}

func freshFields() map[string]string {
	return map[string]string{
		"auth_date": strconv.FormatInt(time.Now().Unix(), 10),
		"query_id":  "AAEtest",
		"user":      `{"id":12345,"first_name":"Иван","username":"ivan"}`,
	}
}

func TestValidateInitData_AcceptsValidSignature(t *testing.T) {
	initData := signValues(testBotToken, freshFields()).Encode()

	user, err := ValidateInitData(initData, testBotToken, time.Hour)
	if err != nil {
		t.Fatalf("expected valid init data to pass, got %v", err)
	}

	if user.ID != 12345 {
		t.Errorf("expected user id 12345, got %d", user.ID)
	}
}

func TestValidateInitData_RejectsTamperedField(t *testing.T) {
	values := signValues(testBotToken, freshFields())
	// Меняем user уже после подписи — data_check_string не сойдётся с hash.
	values.Set("user", `{"id":99999,"first_name":"Мария"}`)

	_, err := ValidateInitData(values.Encode(), testBotToken, time.Hour)
	if err == nil {
		t.Fatal("expected tampered init data to be rejected")
	}
}

func TestValidateInitData_RejectsWrongToken(t *testing.T) {
	initData := signValues(testBotToken, freshFields()).Encode()

	_, err := ValidateInitData(initData, "999:other-bot-token", time.Hour)
	if err == nil {
		t.Fatal("expected init data signed by another token to be rejected")
	}
}

func TestValidateInitData_RejectsExpired(t *testing.T) {
	fields := freshFields()
	fields["auth_date"] = strconv.FormatInt(time.Now().Add(-48*time.Hour).Unix(), 10)
	initData := signValues(testBotToken, fields).Encode()

	_, err := ValidateInitData(initData, testBotToken, 24*time.Hour)
	if err == nil {
		t.Fatal("expected stale init data to be rejected")
	}
}

func TestValidateInitData_RejectsEmptyToken(t *testing.T) {
	initData := signValues(testBotToken, freshFields()).Encode()

	// Пустой токен: без секрета подделку не отличить — отказываем (fail closed).
	_, err := ValidateInitData(initData, "", time.Hour)
	if err == nil {
		t.Fatal("expected empty bot token to be rejected")
	}
}
