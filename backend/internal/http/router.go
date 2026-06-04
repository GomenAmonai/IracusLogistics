package http

import (
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
	JWTSecret       string
}

func NewRouter(deps RouterDeps) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	health := handlers.NewHealthHandler(deps.DB)
	lead := handlers.NewLeadHandler(deps.LeadService)
	auth := handlers.NewAuthHandler(deps.AuthService)
	clientAuth := handlers.NewClientAuthHandler(deps.ClientService)
	client := handlers.NewClientHandler(deps.ClientService)
	shipment := handlers.NewShipmentHandler(deps.ShipmentService, deps.MessageService)
	appShipment := handlers.NewAppShipmentHandler(deps.ShipmentService, deps.MessageService)

	api := router.Group("/api")
	{
		// Публичные ручки.
		api.GET("/health", health.Handle)
		api.POST("/leads", lead.Create)                     // форма с сайта, без авторизации
		api.POST("/auth/login", auth.Login)                 // менеджер: email + пароль
		api.POST("/app/auth/telegram", clientAuth.Telegram) // клиент: Telegram initData

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

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
