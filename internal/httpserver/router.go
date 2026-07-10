package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/teaspeak-v2/wt-bot-ms-payments-v1/docs"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/handlers"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/httpserver/middleware"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/token"
)

type RouterDeps struct {
	Plans        *handlers.PlanHandler
	Subscriptions *handlers.SubscriptionHandler
	Payments     *handlers.PaymentHandler
	Health       *handlers.HealthHandler
	Tokens       *token.Manager
	ServiceAPIKey string
	AllowedOrigins []string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recover(nil), middleware.Logger(nil), middleware.CORS(deps.AllowedOrigins))
	r.Get("/", deps.Health.Root)
	r.Get("/healthz", deps.Health.Healthz)
	r.Get("/readyz", deps.Health.Readyz)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(auth chi.Router) {
			auth.Use(middleware.Auth(deps.Tokens))

			auth.Get("/plans", deps.Plans.List)
			auth.Post("/plans", deps.Plans.Create)
			auth.Get("/plans/{id}", deps.Plans.GetByID)
			auth.Patch("/plans/{id}", deps.Plans.Update)
			auth.Delete("/plans/{id}", deps.Plans.Delete)

			auth.Get("/subscriptions", deps.Subscriptions.List)
			auth.Post("/subscriptions", deps.Subscriptions.Create)
			auth.Get("/subscriptions/{id}", deps.Subscriptions.GetByID)
			auth.Patch("/subscriptions/{id}", deps.Subscriptions.Update)
			auth.Delete("/subscriptions/{id}", deps.Subscriptions.Delete)

			auth.Get("/payments", deps.Payments.List)
			auth.Post("/payments", deps.Payments.Create)
			auth.Get("/payments/{id}", deps.Payments.GetByID)
			auth.Patch("/payments/{id}", deps.Payments.Update)
			auth.Delete("/payments/{id}", deps.Payments.Delete)
		})

		api.Group(func(svc chi.Router) {
			svc.Use(middleware.ServiceKey(deps.ServiceAPIKey))

			svc.Get("/subscriptions/{id}/status", deps.Subscriptions.Status)
			svc.Post("/payments/{id}/status", deps.Payments.UpdateStatus)
		})

		api.Group(func(admin chi.Router) {
			admin.Use(middleware.Auth(deps.Tokens), middleware.RequireRole(models.RoleAdmin))

			_ = admin
		})
	})

	return r
}
