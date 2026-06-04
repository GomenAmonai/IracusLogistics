package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/middleware"
	"icaris-logistic/backend/internal/token"
)

const secret = "test-secret"

func managerToken(t *testing.T, id uuid.UUID) string {
	t.Helper()
	signed, err := token.Issue([]byte(secret), id, domain.RoleManager, time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return signed
}

func clientToken(t *testing.T, id uuid.UUID) string {
	t.Helper()
	signed, err := token.Issue([]byte(secret), id, domain.RoleClient, time.Hour)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return signed
}

func serve(mw gin.HandlerFunc, authHeader string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	engine.Use(mw)
	engine.GET("/x", handler)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	engine.ServeHTTP(rec, req)
	return rec
}

func serveWithAuth(authHeader string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	return serve(middleware.RequireAuth(secret), authHeader, handler)
}

func okHandler(c *gin.Context) { c.Status(http.StatusOK) }

func TestRequireAuth_MissingTokenReturns401(t *testing.T) {
	rec := serveWithAuth("", okHandler)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidTokenReturns401(t *testing.T) {
	rec := serveWithAuth("Bearer not-a-jwt", okHandler)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidTokenPasses(t *testing.T) {
	rec := serveWithAuth("Bearer "+managerToken(t, uuid.New()), okHandler)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_TokenWithoutExpiryReturns401(t *testing.T) {
	// Корректно подписан, но без exp — WithExpirationRequired должен его отвергнуть.
	claims := jwt.RegisteredClaims{Subject: uuid.New().String()}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	rec := serveWithAuth("Bearer "+signed, okHandler)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for token without exp, got %d", rec.Code)
	}
}

func TestRequireAuth_PutsManagerIDFromSubjectIntoContext(t *testing.T) {
	id := uuid.New()
	var got uuid.UUID
	var ok bool

	serveWithAuth("Bearer "+managerToken(t, id), func(c *gin.Context) {
		got, ok = middleware.ManagerID(c)
		c.Status(http.StatusOK)
	})

	if !ok || got != id {
		t.Errorf("expected manager id %s in context, got %s (ok=%v)", id, got, ok)
	}
}

func TestRequireAuth_RejectsClientRoleToken(t *testing.T) {
	// Клиентский токен валидно подписан, но role=client — менеджерская проверка его не пускает.
	rec := serveWithAuth("Bearer "+clientToken(t, uuid.New()), okHandler)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for client-role token on manager middleware, got %d", rec.Code)
	}
}

func TestRequireClientAuth_ValidClientTokenPasses(t *testing.T) {
	rec := serve(middleware.RequireClientAuth(secret), "Bearer "+clientToken(t, uuid.New()), okHandler)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireClientAuth_RejectsManagerRoleToken(t *testing.T) {
	rec := serve(middleware.RequireClientAuth(secret), "Bearer "+managerToken(t, uuid.New()), okHandler)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for manager-role token on client middleware, got %d", rec.Code)
	}
}

func TestRequireClientAuth_PutsClientIDFromSubjectIntoContext(t *testing.T) {
	id := uuid.New()
	var got uuid.UUID
	var ok bool

	serve(middleware.RequireClientAuth(secret), "Bearer "+clientToken(t, id), func(c *gin.Context) {
		got, ok = middleware.ClientID(c)
		c.Status(http.StatusOK)
	})

	if !ok || got != id {
		t.Errorf("expected client id %s in context, got %s (ok=%v)", id, got, ok)
	}
}
