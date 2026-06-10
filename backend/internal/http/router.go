package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"icaris-logistic/backend/internal/http/handlers"
	"icaris-logistic/backend/internal/middleware"
	"icaris-logistic/backend/internal/service"
)

type RouterDeps struct {
	DB              *gorm.DB
	LeadService     *service.LeadService
	AuthService     *service.AuthService
	ClientService   *service.ClientService
	ShipmentService *service.ShipmentService
	MessageService  *service.MessageService
	PaymentService  *service.PaymentService
	JWTSecret       string

	// AllowedOrigins — белый список CORS. AllowAnyOrigin (только dev) отдаёт «*» и игнорирует
	// список. ReleaseMode переводит Gin в прод-режим (без debug-логов).
	AllowedOrigins []string
	AllowAnyOrigin bool
	ReleaseMode    bool

	// TrustedProxies — IP/CIDR прокси, чьим X-Forwarded-For верит ClientIP (а значит и
	// rate-limit). Пусто => приватные диапазоны: на Railway/в докере до контейнера
	// дотягивается только платформенный балансировщик, а публичный клиент с приватного
	// адреса прийти не может, подделать заголовок снаружи нельзя (техдолг #20).
	TrustedProxies []string
}

// defaultTrustedProxies — приватные сети (RFC1918), CGNAT (Railway) и loopback.
var defaultTrustedProxies = []string{
	"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10", "127.0.0.1", "::1",
}

// NewRouter собирает роутер. Ошибка возможна только на битом TRUSTED_PROXIES — это
// ошибка конфигурации, по ней процесс должен упасть на старте (fail-fast, как Validate).
func NewRouter(deps RouterDeps) (*gin.Engine, error) {
	if deps.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(deps.AllowedOrigins, deps.AllowAnyOrigin))

	trusted := deps.TrustedProxies
	if len(trusted) == 0 {
		trusted = defaultTrustedProxies
	}
	if err := router.SetTrustedProxies(trusted); err != nil {
		return nil, fmt.Errorf("router: set trusted proxies: %w", err)
	}

	health := handlers.NewHealthHandler(deps.DB)
	lead := handlers.NewLeadHandler(deps.LeadService)
	auth := handlers.NewAuthHandler(deps.AuthService)
	clientAuth := handlers.NewClientAuthHandler(deps.ClientService)
	client := handlers.NewClientHandler(deps.ClientService)
	shipment := handlers.NewShipmentHandler(deps.ShipmentService, deps.MessageService)
	appShipment := handlers.NewAppShipmentHandler(deps.ShipmentService, deps.MessageService)
	payment := handlers.NewPaymentHandler(deps.PaymentService)

	api := router.Group("/api")
	{
		// Публичные ручки. Свой лимитер на каждую: общий бакет дал бы спаму лидами
		// выбить логин с того же IP (NAT/прокси). Защита от спама и перебора.
		api.GET("/health", health.Handle)
		api.POST("/leads", middleware.RateLimit(1, 5), lead.Create)                     // форма с сайта, без авторизации
		api.POST("/auth/login", middleware.RateLimit(1, 5), auth.Login)                 // менеджер: email + пароль
		api.POST("/app/auth/telegram", middleware.RateLimit(1, 5), clientAuth.Telegram) // клиент: Telegram initData

		// Менеджерские ручки: валидный JWT с role=manager.
		manager := api.Group("")
		manager.Use(middleware.RequireAuth(deps.JWTSecret))
		{
			manager.GET("/leads", lead.List)
			manager.GET("/leads/:id", lead.GetByID)
			manager.PATCH("/leads/:id", lead.UpdateStatus)

			manager.GET("/clients", client.List)

			manager.POST("/shipments", shipment.Create)
			manager.GET("/shipments", shipment.List)
			manager.GET("/shipments/:id", shipment.GetByID)
			manager.PATCH("/shipments/:id/status", shipment.UpdateStatus)
			manager.GET("/shipments/:id/messages", shipment.ListMessages)
			manager.POST("/shipments/:id/messages", shipment.SendMessage)

			manager.POST("/shipments/:id/payments", payment.Create)
			manager.GET("/shipments/:id/payments", payment.List)
			manager.PATCH("/shipments/:id/payments/:paymentID", payment.UpdateStatus)
		}

		// Клиентские ручки WebApp: валидный JWT с role=client.
		app := api.Group("/app")
		app.Use(middleware.RequireClientAuth(deps.JWTSecret))
		{
			app.GET("/shipments", appShipment.List)
			app.GET("/shipments/:id", appShipment.GetByID)
			app.GET("/shipments/:id/messages", appShipment.ListMessages)
			app.POST("/shipments/:id/messages", appShipment.SendMessage)
		}
	}

	return router, nil
}

// corsMiddleware отдаёт CORS-заголовки. allowAny (только dev) разрешает любой origin через «*».
// Иначе Access-Control-Allow-Origin выставляется, только если Origin запроса в белом списке —
// чужие источники браузер заблокирует. Пустой список вне dev => запрещены все (deny by default).
func corsMiddleware(allowedOrigins []string, allowAny bool) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		if allowAny {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			// Ответ зависит от Origin (эхо только для белого списка) — помечаем Vary даже когда
			// origin не совпал, иначе общий кэш мог бы отдать без-CORS ответ доверенному origin.
			c.Header("Vary", "Origin")
			if origin := c.GetHeader("Origin"); origin != "" {
				if _, ok := allowed[origin]; ok {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
