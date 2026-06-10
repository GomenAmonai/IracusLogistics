package http_test

import (
	"testing"

	apphttp "icaris-logistic/backend/internal/http"
)

func TestNewRouterRejectsInvalidTrustedProxy(t *testing.T) {
	_, err := apphttp.NewRouter(apphttp.RouterDeps{TrustedProxies: []string{"not-a-cidr"}})

	if err == nil {
		t.Error("expected error for invalid trusted proxy value")
	}
}

func TestNewRouterDefaultsTrustedProxies(t *testing.T) {
	_, err := apphttp.NewRouter(apphttp.RouterDeps{})

	if err != nil {
		t.Errorf("expected router to build with default trusted proxies, got %v", err)
	}
}
