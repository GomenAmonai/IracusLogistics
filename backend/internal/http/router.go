package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"iracus-logistic/backend/internal/http/handlers"
	"iracus-logistic/backend/internal/middleware"
	"iracus-logistic/backend/internal/service"
)

type RouterDeps struct {
	DB          *gorm.DB
	LeadService *service.LeadService
	AuthService *service.AuthService
	JWTSecret   string
}

func NewRouter(deps RouterDeps) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	health := handlers.NewHealthHandler(deps.DB)
	lead := handlers.NewLeadHandler(deps.LeadService)
	auth := handlers.NewAuthHandler(deps.AuthService)

	api := router.Group("/api")
	{
		// Публичные ручки.
		api.GET("/health", health.Handle)
		api.POST("/leads", lead.Create) // форма с сайта, без авторизации
		api.POST("/auth/login", auth.Login)

		// Защищённые: только менеджер с валидным JWT.
		protected := api.Group("")
		protected.Use(middleware.RequireAuth(deps.JWTSecret))
		{
			protected.GET("/leads", lead.List)
			protected.GET("/leads/:id", lead.GetByID)
			protected.PATCH("/leads/:id", lead.UpdateStatus)
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
