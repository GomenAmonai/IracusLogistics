package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"iracus-logistic/backend/internal/http/handlers"
	"iracus-logistic/backend/internal/shipment"
)

type RouterDeps struct {
	DB              *pgxpool.Pool
	ShipmentService shipment.ServiceAPI
}

func NewRouter(deps RouterDeps) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware)

	healthHandler := handlers.NewHealthHandler(deps.DB)
	shipmentHandler := handlers.NewShipmentRequestHandler(deps.ShipmentService)

	router.Route("/api", func(router chi.Router) {
		router.Get("/health", healthHandler.Handle)
		router.Route("/shipment-requests", func(router chi.Router) {
			router.Get("/", shipmentHandler.List)
			router.Post("/", shipmentHandler.Create)
			router.Route("/{id}", func(router chi.Router) {
				router.Get("/", shipmentHandler.Get)
				router.Patch("/", shipmentHandler.Update)
			})
		})
	})

	return router
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
